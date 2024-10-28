package worker

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"sync"
	"testing"
	"time"
)

type mockTask struct{}
type mockTaskResult struct{}

type mockTaskProcessor struct {
	numTasks int
	process  func(ctx context.Context, task *mockTask) (*mockTaskResult, error)
	complete func(ctx context.Context, result *mockTaskResult) error
	abandon  func(ctx context.Context, task *mockTask, reason *string) error
	sync.Mutex
}

func (m *mockTaskProcessor) GetTask(_ context.Context) (*mockTask, error) {
	m.Lock()
	defer m.Unlock()
	if m.numTasks > 0 {
		m.numTasks--
		return &mockTask{}, nil
	} else {
		return nil, ErrNoTask
	}
}

func (m *mockTaskProcessor) ProcessTask(ctx context.Context, task *mockTask) (*mockTaskResult, error) {
	return m.process(ctx, task)
}

func (m *mockTaskProcessor) CompleteTask(ctx context.Context, result *mockTaskResult) error {
	return m.complete(ctx, result)
}

func (m *mockTaskProcessor) AbandonTask(ctx context.Context, task *mockTask, reason *string) error {
	return m.abandon(ctx, task, reason)
}

func TestWorkerDraining(t *testing.T) {
	taskProcessor := &mockTaskProcessor{
		numTasks: 10,
		process: func(ctx context.Context, task *mockTask) (*mockTaskResult, error) {
			time.Sleep(2 * time.Second)
			return &mockTaskResult{}, nil
		},
		complete: func(ctx context.Context, result *mockTaskResult) error {
			return nil
		},
		abandon: func(ctx context.Context, task *mockTask, reason *string) error {
			return nil
		},
	}
	w := NewWorker("worker", taskProcessor, zap.NewNop(), WithMaxConcurrentTasksLimit(3))
	ctx := context.Background()
	w.Start(ctx)
	time.Sleep(10 * time.Second)
	w.Stop(ctx)
}

func TestProcessError(t *testing.T) {
	taskProcessor := &mockTaskProcessor{
		numTasks: 10,
		process: func(ctx context.Context, task *mockTask) (*mockTaskResult, error) {
			return nil, errors.New("always error")
		},
		complete: func(ctx context.Context, result *mockTaskResult) error {
			return nil
		},
		abandon: func(ctx context.Context, task *mockTask, reason *string) error {
			fmt.Printf("%v\n", reason)
			return nil
		},
	}
	w := NewWorker("worker", taskProcessor, zap.NewNop(), WithMaxConcurrentTasksLimit(1))
	ctx := context.Background()
	w.Start(ctx)
	time.Sleep(5 * time.Second)
	w.Stop(ctx)
}

func TestPanicProcessor(t *testing.T) {
	taskProcessor := &mockTaskProcessor{
		numTasks: 5,
		process: func(ctx context.Context, task *mockTask) (*mockTaskResult, error) {
			return &mockTaskResult{}, nil
		},
		complete: func(ctx context.Context, result *mockTaskResult) error {
			panic("panicked")
		},
		abandon: func(ctx context.Context, task *mockTask, reason *string) error {
			panic("panicked")
		},
	}
	w := NewWorker("worker", taskProcessor, zap.NewNop(), WithMaxConcurrentTasksLimit(1))
	ctx := context.Background()
	w.Start(ctx)
	time.Sleep(3 * time.Second)
	w.Stop(ctx)
	//
	taskProcessor.process = func(ctx context.Context, task *mockTask) (*mockTaskResult, error) {
		return nil, errors.New("always error")
	}
	w = NewWorker("worker", taskProcessor, zap.NewNop(), WithMaxConcurrentTasksLimit(1))
	ctx = context.Background()
	w.Start(ctx)
	time.Sleep(3 * time.Second)
	w.Stop(ctx)
}
