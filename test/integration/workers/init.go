package workers

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflows-go/pkg/api"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
	"github.com/tuannh982/simple-workflows-go/pkg/worker"
	"github.com/tuannh982/simple-workflows-go/test/integration/mocks"
	"testing"
	"time"
)

type mockStruct struct{}

func mockActivity1(_ context.Context, input *mockStruct) (*mockStruct, error) { panic("mock") }
func mockActivity2(_ context.Context, input *mockStruct) (*mockStruct, error) { panic("mock") }

func mockWorkflow1(ctx context.Context, input *mockStruct) (*mockStruct, error) {
	var r *mockStruct
	var err error
	r, err = api.CallActivity(ctx, mockActivity1, input).Await()
	_, err = api.CreateTimer(ctx, 50*time.Second).Await()
	r, err = api.CallActivity(ctx, mockActivity2, input).Await()
	return r, err
}

var dataConverter = dataconverter.NewJsonDataConverter()

func initWorkers(t *testing.T) (*worker.ActivityWorker, *worker.WorkflowWorker) {
	var err error
	be := mocks.NewMockBackend()
	ar := registry.NewActivityRegistry()
	err = ar.RegisterActivities(
		mockActivity1,
		mockActivity2,
	)
	assert.Nil(t, err)
	wr := registry.NewWorkflowRegistry()
	err = wr.RegisterWorkflows(mockWorkflow1)
	assert.Nil(t, err)
	activityWorker := worker.NewActivityWorker(be, ar, dataConverter)
	workflowWorker := worker.NewWorkflowWorker(be, wr, dataConverter)
	return activityWorker, workflowWorker
}
