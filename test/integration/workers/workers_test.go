package workers

import (
	"context"
	"testing"
	"time"
)

func TestWorkers(t *testing.T) {
	activityWorker, workflowWorker := initWorkers(t)
	// TODO implement me
	ctx := context.Background()
	activityWorker.Start(ctx)
	workflowWorker.Start(ctx)
	time.Sleep(5 * time.Second)
	activityWorker.Stop(ctx)
	workflowWorker.Stop(ctx)
}
