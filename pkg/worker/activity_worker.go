package worker

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/internal/activity"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/worker"
)

type ActivityWorker struct {
	processor worker.TaskProcessor[task.ActivityTask, task.ActivityTaskResult]
	w         worker.Worker[task.ActivityTask, task.ActivityTaskResult]
}

func NewActivityWorker(
	be backend.Backend,
	registry *registry.ActivityRegistry,
	dataConverter dataconverter.DataConverter,
	opts ...func(options *worker.WorkerOptions),
) *ActivityWorker {
	executor := activity.NewActivityTaskExecutor(registry, dataConverter)
	processor := activity.NewActivityTaskProcessor(be, executor)
	w := worker.NewWorker("activity worker", processor, opts...)
	return &ActivityWorker{
		processor: processor,
		w:         w,
	}
}

func (a *ActivityWorker) Start(ctx context.Context) {
	a.w.Start(ctx)
}

func (a *ActivityWorker) Stop(ctx context.Context) {
	a.w.Stop(ctx)
}
