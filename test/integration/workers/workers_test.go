package workers

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflows-go/pkg/api"
	"testing"
	"time"
)

func TestWorkers(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	be, activityWorker, workflowWorker := initWorkers(t)
	// TODO implement me
	ctx := context.Background()
	_ = activityWorker
	activityWorker.Start(ctx)
	workflowWorker.Start(ctx)
	err := api.ScheduleWorkflow(dataConverter, be, mockWorkflow1, &mockStruct{})
	assert.NoError(t, err)
	time.Sleep(15 * time.Second)
	err = be.GetWorkflowResult(ctx, "test")
	assert.NoError(t, err)
	activityWorker.Stop(ctx)
	workflowWorker.Stop(ctx)
}
