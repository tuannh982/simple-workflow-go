package worker

import (
	"context"
	"testing"
	"time"
)

type mockTask struct{}
type mockTaskResult struct{}

type mockTaskProcessor struct {
	numTasks int
}

func (m *mockTaskProcessor) GetTask(ctx context.Context) (*mockTask, error) {
	if m.numTasks > 0 {
		m.numTasks--
		return &mockTask{}, nil
	} else {
		return nil, ErrNoTask
	}
}

func (m *mockTaskProcessor) ProcessTask(ctx context.Context, task *mockTask) (*mockTaskResult, error) {
	time.Sleep(2 * time.Second)
	return &mockTaskResult{}, nil
}

func (m *mockTaskProcessor) CompleteTask(ctx context.Context, result *mockTaskResult) error {
	return nil
}

func (m *mockTaskProcessor) AbandonTask(ctx context.Context, task *mockTask) error {
	return nil
}

func TestWorkerDrainingStop(t *testing.T) {
	taskProcessor := &mockTaskProcessor{
		numTasks: 10,
	}
	worker := NewWorker("worker", taskProcessor, WithMaxConcurrentTasksLimit(3))
	ctx := context.Background()
	worker.Start(ctx)
	time.Sleep(10 * time.Second)
	worker.Stop(ctx)
}
