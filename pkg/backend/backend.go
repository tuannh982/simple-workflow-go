package backend

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/dto"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
)

type Backend interface {
	DataConverter() dataconverter.DataConverter
	CreateWorkflow(ctx context.Context, info *history.WorkflowExecutionStarted) error
	GetWorkflowResult(ctx context.Context, name string, workflowID string) (*dto.WorkflowExecutionResult, error)
	AppendWorkflowEvent(ctx context.Context, workflowID string, event *history.HistoryEvent) error
	GetWorkflowHistory(ctx context.Context, workflowID string) ([]*history.HistoryEvent, error)
	GetWorkflowTask(ctx context.Context) (*task.WorkflowTask, error)
	CompleteWorkflowTask(ctx context.Context, result *task.WorkflowTaskResult) error
	AbandonWorkflowTask(ctx context.Context, task *task.WorkflowTask, reason *string) error
	GetActivityTask(ctx context.Context) (*task.ActivityTask, error)
	CompleteActivityTask(ctx context.Context, result *task.ActivityTaskResult) error
	AbandonActivityTask(ctx context.Context, task *task.ActivityTask, reason *string) error
}
