package activity

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/worker"
	"go.uber.org/zap"
)

type activityTaskProcessor struct {
	be       backend.Backend
	executor ActivityTaskExecutor
	logger   *zap.Logger
}

func NewActivityTaskProcessor(
	be backend.Backend,
	executor ActivityTaskExecutor,
	_ *zap.Logger,
) worker.TaskProcessor[*task.ActivityTask, *task.ActivityTaskResult] {
	return &activityTaskProcessor{
		be:       be,
		executor: executor,
	}
}

func (a *activityTaskProcessor) GetTask(ctx context.Context) (*task.ActivityTask, error) {
	return a.be.GetActivityTask(ctx)
}

func (a *activityTaskProcessor) ProcessTask(ctx context.Context, task *task.ActivityTask) (*task.ActivityTaskResult, error) {
	return a.executor.Execute(ctx, task)
}

func (a *activityTaskProcessor) CompleteTask(ctx context.Context, result *task.ActivityTaskResult) error {
	return a.be.CompleteActivityTask(ctx, result)
}

func (a *activityTaskProcessor) AbandonTask(ctx context.Context, task *task.ActivityTask, reason *string) error {
	return a.be.AbandonActivityTask(ctx, task, reason)
}
