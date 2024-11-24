package workflow_task_executor

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflow-go/internal/workflow"
	workflow2 "github.com/tuannh982/simple-workflow-go/pkg/api/workflow"
	"github.com/tuannh982/simple-workflow-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflow-go/pkg/registry"
	"go.uber.org/zap"
	"testing"
	"time"
)

type mockStruct struct{}

func mockActivity1(_ context.Context, _ *mockStruct) (*mockStruct, error) { panic("mock") }
func mockActivity2(_ context.Context, _ *mockStruct) (*mockStruct, error) { panic("mock") }

func mockWorkflow1(ctx context.Context, input *mockStruct) (*mockStruct, error) {
	var r *mockStruct
	var err error
	r, err = workflow2.CallActivity(ctx, mockActivity1, input).Await()
	workflow2.WaitFor(ctx, 50*time.Second)
	r, err = workflow2.CallActivity(ctx, mockActivity2, input).Await()
	return r, err
}

var dataConverter = dataconverter.NewJsonDataConverter()

func initExecutor(t *testing.T) workflow.WorkflowTaskExecutor {
	var err error
	r := registry.NewWorkflowRegistry()
	err = r.RegisterWorkflows(mockWorkflow1)
	assert.Nil(t, err)
	executor := workflow.NewWorkflowTaskExecutor(r, dataConverter, zap.NewNop())
	return executor
}
