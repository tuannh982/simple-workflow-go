package backend

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/internal/dto/history"
	"github.com/tuannh982/simple-workflows-go/internal/dto/task"
)

type Backend interface {
	Start(context.Context) error
	Stop(context.Context) error
	CreateWorkflow(ctx context.Context, info *history.WorkflowExecutionStarted) error
	AppendWorkflowHistory(context.Context, string, *history.HistoryEvent) error
	GetWorkflowTask(ctx context.Context) (*task.WorkflowTask, error)
	CompleteWorkflowTask(ctx context.Context, result *task.WorkflowTaskResult) error
	AbandonWorkflowTask(ctx context.Context, task *task.WorkflowTask) error
	GetActivityTask(ctx context.Context) (*task.ActivityTask, error)
	CompleteActivityTask(ctx context.Context, result *task.ActivityTaskResult) error
	AbandonActivityTask(ctx context.Context, task *task.ActivityTask) error
}
