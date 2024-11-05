package workers

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflows-go/pkg/api/workflow"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
	"github.com/tuannh982/simple-workflows-go/pkg/worker/activity_worker"
	"github.com/tuannh982/simple-workflows-go/pkg/worker/workflow_worker"
	"github.com/tuannh982/simple-workflows-go/test/integration/mocks"
	"go.uber.org/zap"
	"testing"
	"time"
)

type mockStruct struct {
	Msg string
}

func mockActivity1(_ context.Context, input *mockStruct) (*mockStruct, error) {
	return &mockStruct{Msg: fmt.Sprintf("%s,activity_1", input.Msg)}, nil
}

func mockActivity2(_ context.Context, input *mockStruct) (*mockStruct, error) {
	return &mockStruct{Msg: fmt.Sprintf("%s,activity_2", input.Msg)}, nil
}

func mockWorkflow1(ctx context.Context, input *mockStruct) (r *mockStruct, err error) {
	r = input
	r, err = workflow.CallActivity(ctx, mockActivity1, r).Await()
	if err != nil {
		return nil, err
	}
	workflow.SetVar(ctx, "activity_1_result", r)
	workflow.WaitFor(ctx, 1*time.Hour)
	r, err = workflow.CallActivity(ctx, mockActivity2, r).Await()
	if err != nil {
		return nil, err
	}
	return r, err
}

var dataConverter = dataconverter.NewJsonDataConverter()

func initWorkers(t *testing.T, logger *zap.Logger) (backend.Backend, *activity_worker.ActivityWorker, *workflow_worker.WorkflowWorker) {
	var err error
	be := mocks.NewMockBackend(dataConverter)
	ar := registry.NewActivityRegistry()
	err = ar.RegisterActivities(
		mockActivity1,
		mockActivity2,
	)
	assert.Nil(t, err)
	wr := registry.NewWorkflowRegistry()
	err = wr.RegisterWorkflows(mockWorkflow1)
	assert.Nil(t, err)
	activityWorker := activity_worker.NewActivityWorker("1", be, ar, dataConverter, logger)
	workflowWorker := workflow_worker.NewWorkflowWorker("1", be, wr, dataConverter, logger)
	return be, activityWorker, workflowWorker
}
