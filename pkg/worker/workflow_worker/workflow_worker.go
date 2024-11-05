package workflow_worker

import (
	"context"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/internal/workflow"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/worker"
	"go.uber.org/zap"
)

type WorkflowWorker struct {
	name      string
	processor worker.TaskProcessor[*task.WorkflowTask, *task.WorkflowTaskResult]
	w         worker.Worker[*task.WorkflowTask, *task.WorkflowTaskResult]
	logger    *zap.Logger
}

func NewWorkflowWorker(
	name string,
	be backend.Backend,
	registry *registry.WorkflowRegistry,
	dataConverter dataconverter.DataConverter,
	logger *zap.Logger,
	opts ...func(options *WorkflowWorkerOptions),
) *WorkflowWorker {
	options := NewWorkflowWorkerOptions()
	for _, configure := range opts {
		configure(options)
	}
	fqn := fmt.Sprintf("Activity worker %s", name)
	childLogger := logger.With(zap.String("worker", name))
	executor := workflow.NewWorkflowTaskExecutor(registry, dataConverter, childLogger)
	processor := workflow.NewWorkflowTaskProcessor(be, executor, childLogger)
	w := worker.NewWorker(fqn, processor, childLogger, options.WorkerOptions)
	return &WorkflowWorker{
		processor: processor,
		w:         w,
		logger:    childLogger,
	}
}

func (w *WorkflowWorker) Start(ctx context.Context) {
	w.w.Start(ctx)
}

func (w *WorkflowWorker) Stop(ctx context.Context) {
	w.w.Stop(ctx)
}
