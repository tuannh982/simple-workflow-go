package workers

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflows-go/pkg/api/client"
	"testing"
)

func TestWorkers(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	be, activityWorker, workflowWorker := initWorkers(t)
	ctx := context.Background()
	activityWorker.Start(ctx)
	defer activityWorker.Stop(ctx)
	workflowWorker.Start(ctx)
	defer workflowWorker.Stop(ctx)
	err := client.ScheduleWorkflow(ctx, be, mockWorkflow1, &mockStruct{Msg: "initial"}, client.WorkflowScheduleOptions{
		WorkflowID: "mock-workflow-id",
		Version:    "1",
	})
	assert.NoError(t, err)
	wResult, wErr, err := client.AwaitWorkflowResult(ctx, be, mockWorkflow1, "mock-workflow-id")
	assert.NoError(t, err)
	assert.Equal(t, "initial,activity_1,activity_2", wResult.Msg)
	assert.Nil(t, wErr)
}
