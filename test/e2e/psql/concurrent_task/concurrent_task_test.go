//go:build e2e
// +build e2e

package concurrent_task

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflow-go/pkg/api/client"
	"github.com/tuannh982/simple-workflow-go/pkg/api/worker"
	"github.com/tuannh982/simple-workflow-go/pkg/worker/activity_worker"
	"github.com/tuannh982/simple-workflow-go/pkg/worker/workflow_worker"
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
		WithActivityWorkerOpts(
			activity_worker.WithMaxConcurrentTasksLimit(20),
			activity_worker.WithPollerInitialBackoffInterval(50*time.Millisecond),
			activity_worker.WithPollerMaxBackoffInterval(100*time.Millisecond),
		).
		RegisterActivities(
			GenerateNumberActivity,
		).
		Build()
	assert.NoError(t, err)
	ww, err := worker.NewWorkflowWorkersBuilder().
		WithName("[e2e test] WorkflowWorker").
		WithBackend(be).
		WithLogger(logger).
		WithWorkflowWorkerOpts(
			workflow_worker.WithMaxConcurrentTasksLimit(1),
			workflow_worker.WithPollerInitialBackoffInterval(1000*time.Millisecond),
		).
		RegisterWorkflows(
			SumWorkflow,
		).
		Build()
	assert.NoError(t, err)
	aw.Start(ctx)
	defer aw.Stop(ctx)
	ww.Start(ctx)
	defer ww.Stop(ctx)
	// mock results
	numTasks := 40
	var expectedResult int64 = 0
	for i := 0; i < numTasks; i++ {
		expectedResult += GenerateNumber(int64(i))
	}
	//
	uid, err := uuid.NewV6()
	assert.NoError(t, err)
	workflowID := fmt.Sprintf("e2e-concurrent-task-workflow-%s", uid.String())
	err = client.ScheduleWorkflow(ctx, be, SumWorkflow, &SumWorkflowInput{
		NumberOfTasks: numTasks,
	}, client.WorkflowScheduleOptions{
		WorkflowID: workflowID,
		Version:    "1",
	})
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
	wResult, wErr, err := client.AwaitWorkflowResult(ctx, be, SumWorkflow, workflowID)
	assert.NotNil(t, wResult)
	assert.NoError(t, wErr)
	//goland:noinspection GoDfaErrorMayBeNotNil
	assert.Equal(t, expectedResult, wResult.Result)
}
