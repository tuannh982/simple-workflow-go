package workers

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"testing"
)

type mockBackend struct {
}

func (m *mockBackend) Start(ctx context.Context) error { return nil }

func (m *mockBackend) Stop(ctx context.Context) error { return nil }

func (m *mockBackend) CreateWorkflow(ctx context.Context, info *history.WorkflowExecutionStarted) error {
	// TODO implement me
	panic("implement me")
}

func (m *mockBackend) AppendWorkflowHistory(ctx context.Context, s string, event *history.HistoryEvent) error {
	// TODO implement me
	panic("implement me")
}

func (m *mockBackend) GetWorkflowTask(ctx context.Context) (*task.WorkflowTask, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockBackend) CompleteWorkflowTask(ctx context.Context, result *task.WorkflowTaskResult) error {
	// TODO implement me
	panic("implement me")
}

func (m *mockBackend) AbandonWorkflowTask(ctx context.Context, task *task.WorkflowTask, reason *string) error {
	// TODO implement me
	panic("implement me")
}

func (m *mockBackend) GetActivityTask(ctx context.Context) (*task.ActivityTask, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockBackend) CompleteActivityTask(ctx context.Context, result *task.ActivityTaskResult) error {
	// TODO implement me
	panic("implement me")
}

func (m *mockBackend) AbandonActivityTask(ctx context.Context, task *task.ActivityTask, reason *string) error {
	// TODO implement me
	panic("implement me")
}

func TestWorkers(t *testing.T) {
	// TODO implement me
}
