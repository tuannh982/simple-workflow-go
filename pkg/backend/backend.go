package backend

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
)

type Backend interface {
	Start(context.Context) error
	Stop(context.Context) error
	CreateWorkflow(ctx context.Context, info *history.WorkflowExecutionStarted) error
	AppendWorkflowEvent(ctx context.Context, workflowID string, event *history.HistoryEvent) error
	GetWorkflowTask(ctx context.Context) (*task.WorkflowTask, error)
	CompleteWorkflowTask(ctx context.Context, result *task.WorkflowTaskResult) error
	AbandonWorkflowTask(ctx context.Context, task *task.WorkflowTask, reason *string) error
	GetActivityTask(ctx context.Context) (*task.ActivityTask, error)
	CompleteActivityTask(ctx context.Context, result *task.ActivityTaskResult) error
	AbandonActivityTask(ctx context.Context, task *task.ActivityTask, reason *string) error
}
