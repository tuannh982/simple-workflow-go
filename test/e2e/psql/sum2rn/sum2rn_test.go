//go:build e2e
// +build e2e

package concurrent_task

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflows-go/pkg/api/client"
	worker2 "github.com/tuannh982/simple-workflows-go/pkg/api/worker"
	"github.com/tuannh982/simple-workflows-go/test/e2e/psql"
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
	aw, err := worker2.NewActivityWorkersBuilder().
		WithName("[e2e test] ActivityWorker").
		WithBackend(be).
		WithLogger(logger).
		RegisterActivities(
			GenerateRandomNumberActivity1,
			GenerateRandomNumberActivity2,
		).
		Build()
	assert.NoError(t, err)
	ww, err := worker2.NewWorkflowWorkersBuilder().
		WithName("[e2e test] WorkflowWorker").
		WithBackend(be).
		WithLogger(logger).
		RegisterWorkflows(
			Sum2RandomNumberWorkflow,
		).
		Build()
	assert.NoError(t, err)
	aw.Start(ctx)
	defer aw.Stop(ctx)
	ww.Start(ctx)
	defer ww.Stop(ctx)
	// mock results
	var seed int64 = 12423745
	expectedResult := GenerateNumber(seed, 19) + GenerateNumber(seed, 23)
	//
	uid, err := uuid.NewV7()
	assert.NoError(t, err)
	workflowID := fmt.Sprintf("e2e-sum2rn-workflow-%s", uid.String())
	err = client.ScheduleWorkflow(ctx, be, Sum2RandomNumberWorkflow, &Seed{
		Value: seed,
	}, client.WorkflowScheduleOptions{
		WorkflowID: workflowID,
		Version:    "1",
	})
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
	wResult, wErr, err := client.AwaitWorkflowResult(ctx, be, Sum2RandomNumberWorkflow, workflowID)
	assert.NotNil(t, wResult)
	assert.NoError(t, wErr)
	//goland:noinspection GoDfaErrorMayBeNotNil
	assert.Equal(t, expectedResult, wResult.Value)
}
