package activity_worker

import (
	"context"
	"fmt"
	"github.com/tuannh982/simple-workflow-go/internal/activity"
	"github.com/tuannh982/simple-workflow-go/pkg/backend"
	"github.com/tuannh982/simple-workflow-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflow-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflow-go/pkg/registry"
	"github.com/tuannh982/simple-workflow-go/pkg/utils/worker"
	"go.uber.org/zap"
)

type ActivityWorker struct {
	name      string
	processor worker.TaskProcessor[*task.ActivityTask, *task.ActivityTaskResult]
	w         worker.Worker
	logger    *zap.Logger
}

func NewActivityWorker(
	name string,
	be backend.Backend,
	registry *registry.ActivityRegistry,
	dataConverter dataconverter.DataConverter,
	logger *zap.Logger,
	opts ...func(options *ActivityWorkerOptions),
) *ActivityWorker {
	options := NewActivityWorkerOptions()
	for _, configure := range opts {
		configure(options)
	}
	fqn := fmt.Sprintf("Activity worker %s", name)
	childLogger := logger.With(zap.String("worker", name))
	executor := activity.NewActivityTaskExecutor(registry, dataConverter, childLogger)
	processor := activity.NewActivityTaskProcessor(be, executor, childLogger, options.ActivityTaskProcessorOptions)
	w := worker.NewWorker(fqn, processor, childLogger, options.WorkerOptions)
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
