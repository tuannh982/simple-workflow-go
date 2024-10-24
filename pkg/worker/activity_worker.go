package worker

import (
	"context"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/internal/activity"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/worker"
	"go.uber.org/zap"
)

type ActivityWorker struct {
	name      string
	processor worker.TaskProcessor[task.ActivityTask, task.ActivityTaskResult]
	w         worker.Worker[task.ActivityTask, task.ActivityTaskResult]
	logger    *zap.Logger
}

func NewActivityWorker(
	name string,
	be backend.Backend,
	registry *registry.ActivityRegistry,
	dataConverter dataconverter.DataConverter,
	logger *zap.Logger,
	opts ...func(options *worker.WorkerOptions),
) *ActivityWorker {
	fqn := fmt.Sprintf("Activity worker %s", name)
	childLogger := logger.With(zap.String("worker", name))
	executor := activity.NewActivityTaskExecutor(registry, dataConverter, childLogger)
	processor := activity.NewActivityTaskProcessor(be, executor, childLogger)
	w := worker.NewWorker(fqn, processor, childLogger, opts...)
	return &ActivityWorker{
		name:      name,
		processor: processor,
		w:         w,
		logger:    childLogger,
	}
}

func (a *ActivityWorker) Start(ctx context.Context) {
	a.w.Start(ctx)
}

func (a *ActivityWorker) Stop(ctx context.Context) {
	a.w.Stop(ctx)
}
