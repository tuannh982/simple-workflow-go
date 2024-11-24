package workers

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflow-go/pkg/api/client"
	"github.com/tuannh982/simple-workflow-go/pkg/api/debug"
	"github.com/tuannh982/simple-workflow-go/pkg/dto"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestDebuggerGetUserDefinedVars(t *testing.T) {
	logger := zap.NewNop()
	be, activityWorker, workflowWorker := initWorkers(t, logger)
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
	time.Sleep(3 * time.Second) // wait for some time to ensure activity 1 is called
	executionResult, err := client.GetWorkflowResult(ctx, be, mockWorkflow1, "mock-workflow-id")
	assert.NotNil(t, executionResult)
	//goland:noinspection GoDfaErrorMayBeNotNil
	assert.Equal(t, executionResult.RuntimeStatus, string(dto.WorkflowRuntimeStatusRunning))
	dbg := debug.NewWorkflowDebugger(be)
	vars, err := dbg.QueryUserDefinedVars(mockWorkflow1, "mock-workflow-id")
	assert.NoError(t, err)
	assert.NotNil(t, vars["activity_1_result"])
	assert.Equal(t, "initial,activity_1", vars["activity_1_result"].(*mockStruct).Msg)
}
