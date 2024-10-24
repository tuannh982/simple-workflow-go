package workers

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflows-go/pkg/api/workflow"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
	"github.com/tuannh982/simple-workflows-go/pkg/worker"
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

func revertMockActivity1(_ context.Context, input *mockStruct) (*mockStruct, error) {
	return &mockStruct{Msg: fmt.Sprintf("%s,revert_activity_1", input.Msg)}, nil
}

func mockActivity2(_ context.Context, input *mockStruct) (*mockStruct, error) {
	return &mockStruct{Msg: fmt.Sprintf("%s,activity_2", input.Msg)}, nil
}

func revertMockActivity2(_ context.Context, input *mockStruct) (*mockStruct, error) {
	return &mockStruct{Msg: fmt.Sprintf("%s,revert_activity_2", input.Msg)}, nil
}

func mockWorkflow1(ctx context.Context, input *mockStruct) (r *mockStruct, err error) {
	r = input
	r, err = workflow.CallActivity(ctx, mockActivity1, r).Await()
	defer func() {
		if err != nil {
			r, err = workflow.CallActivity(ctx, revertMockActivity1, r).Await()
		}
	}()
	workflow.WaitFor(ctx, 2*time.Second)
	r, err = workflow.CallActivity(ctx, mockActivity2, r).Await()
	defer func() {
		if err != nil {
			r, err = workflow.CallActivity(ctx, revertMockActivity2, r).Await()
		}
	}()
	return r, err
}

var dataConverter = dataconverter.NewJsonDataConverter()

func initWorkers(t *testing.T, logger *zap.Logger) (backend.Backend, *worker.ActivityWorker, *worker.WorkflowWorker) {
	var err error
	be := mocks.NewMockBackend(dataConverter)
	ar := registry.NewActivityRegistry()
	err = ar.RegisterActivities(
		mockActivity1,
		revertMockActivity1,
		mockActivity2,
		revertMockActivity2,
	)
	assert.Nil(t, err)
	wr := registry.NewWorkflowRegistry()
	err = wr.RegisterWorkflows(mockWorkflow1)
	assert.Nil(t, err)
	activityWorker := worker.NewActivityWorker("1", be, ar, dataConverter, logger)
	workflowWorker := worker.NewWorkflowWorker("1", be, wr, dataConverter, logger)
	return be, activityWorker, workflowWorker
}
