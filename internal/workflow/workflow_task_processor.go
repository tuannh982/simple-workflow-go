package workflow

import (
	"context"
	"github.com/tuannh982/simple-workflow-go/pkg/backend"
	"github.com/tuannh982/simple-workflow-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflow-go/pkg/utils/worker"
	"go.uber.org/zap"
)

type workflowTaskProcessor struct {
	be       backend.Backend
	executor WorkflowTaskExecutor
	logger   *zap.Logger
}

func NewWorkflowTaskProcessor(
	be backend.Backend,
	executor WorkflowTaskExecutor,
	logger *zap.Logger,
) worker.TaskProcessor[*task.WorkflowTask, *task.WorkflowTaskResult] {
	return &workflowTaskProcessor{
		be:       be,
		executor: executor,
		logger:   logger,
	}
}

func (w *workflowTaskProcessor) GetTask(ctx context.Context) (*task.WorkflowTask, error) {
	return w.be.GetWorkflowTask(ctx)
}

func (w *workflowTaskProcessor) ProcessTask(ctx context.Context, task *task.WorkflowTask) (*task.WorkflowTaskResult, error) {
	return w.executor.Execute(ctx, task)
}

func (w *workflowTaskProcessor) CompleteTask(ctx context.Context, result *task.WorkflowTaskResult) error {
	return w.be.CompleteWorkflowTask(ctx, result)
}

func (w *workflowTaskProcessor) AbandonTask(ctx context.Context, task *task.WorkflowTask, reason *string) error {
	return w.be.AbandonWorkflowTask(ctx, task, reason)
}
