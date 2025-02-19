//go:build e2e
// +build e2e

package handle_external_event

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflow-go/pkg/api/client"
	"github.com/tuannh982/simple-workflow-go/pkg/api/worker"
	"github.com/tuannh982/simple-workflow-go/test/e2e/psql"
	"go.uber.org/zap"
	"testing"
	"time"
)

func InitLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func Test(t *testing.T) {
	ctx := context.Background()
	logger, err := InitLogger()
	assert.NoError(t, err)
	be, err := psql.InitBackend(logger)
	assert.NoError(t, err)
	aw, err := worker.NewActivityWorkersBuilder().
		WithName("[e2e test] ActivityWorker").
		WithBackend(be).
		WithLogger(logger).
		RegisterActivities(
			Activity1,
		).
		Build()
	assert.NoError(t, err)
	ww, err := worker.NewWorkflowWorkersBuilder().
		WithName("[e2e test] WorkflowWorker").
		WithBackend(be).
		WithLogger(logger).
		RegisterWorkflows(
			HandleExternalEventWorkflow,
		).
		Build()
	assert.NoError(t, err)
	aw.Start(ctx)
	defer aw.Stop(ctx)
	ww.Start(ctx)
	defer ww.Stop(ctx)
	// mock results
	message := "world"
	expectedResult := fmt.Sprintf("Hello, %s!", message)
	//
	uid, err := uuid.NewV6()
	assert.NoError(t, err)
	workflowID := fmt.Sprintf("e2e-external-event-workflow-%s", uid.String())
	err = client.ScheduleWorkflow(ctx, be, HandleExternalEventWorkflow, &Void{}, client.WorkflowScheduleOptions{
		WorkflowID: workflowID,
		Version:    "1",
	})
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
	err = client.SendWorkflowEvent(ctx, be, workflowID, HelloEventName, []byte(message))
	assert.NoError(t, err)
	wResult, wErr, err := client.AwaitWorkflowResult(ctx, be, HandleExternalEventWorkflow, workflowID)
	assert.NotNil(t, wResult)
	assert.NoError(t, wErr)
	//goland:noinspection GoDfaErrorMayBeNotNil
	assert.Equal(t, expectedResult, wResult.Value)
}
