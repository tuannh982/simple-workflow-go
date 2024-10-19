package worker

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/internal/workflow"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/registry"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/worker"
)

type WorkflowWorker struct {
	processor worker.TaskProcessor[task.WorkflowTask, task.WorkflowTaskResult]
	w         worker.Worker[task.WorkflowTask, task.WorkflowTaskResult]
}

func NewWorkflowWorker(
	be backend.Backend,
	registry *registry.WorkflowRegistry,
	dataConverter dataconverter.DataConverter,
	opts ...func(options *worker.WorkerOptions),
) *WorkflowWorker {
	executor := workflow.NewWorkflowTaskExecutor(registry, dataConverter)
	processor := workflow.NewWorkflowTaskProcessor(be, executor)
	w := worker.NewWorker("workflow worker", processor, opts...)
	return &WorkflowWorker{
		processor: processor,
		w:         w,
	}
}

func (w *WorkflowWorker) Start(ctx context.Context) {
	w.w.Start(ctx)
}

func (w *WorkflowWorker) Stop(ctx context.Context) {
	w.w.Stop(ctx)
}
