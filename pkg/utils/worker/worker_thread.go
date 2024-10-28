package worker

import (
	"context"
	"errors"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/backoff"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/ptr"
	"go.uber.org/zap"
	"runtime/debug"
	"sync"
	"time"
)

type WorkerThread[T any, R any] struct {
	name          string
	taskProcessor TaskProcessor[T, R]
	bo            backoff.BackOff
	wg            *sync.WaitGroup
	logger        *zap.Logger
}

func NewWorkerThread[T any, R any](
	name string,
	taskProcessor TaskProcessor[T, R],
	bo backoff.BackOff,
	wg *sync.WaitGroup,
	logger *zap.Logger,
) *WorkerThread[T, R] {
	return &WorkerThread[T, R]{
		name:          name,
		taskProcessor: taskProcessor,
		bo:            bo,
		wg:            wg,
		logger:        logger,
	}
}

func (w *WorkerThread[T, R]) Run(ctx context.Context) {
	defer w.wg.Done()
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			task, err := w.taskProcessor.GetTask(ctx)
			if err != nil {
				if errors.Is(err, ErrNoTask) {
					// expected error, do not log
				} else {
					w.logger.Error("Error while polling task", zap.Error(err))
				}
				w.bo.BackOff()
			} else {
				w.logger.Debug("Task fetched", zap.Any("task", task))
				if err = w.processTask(ctx, task); err != nil {
					w.logger.Error("Error while processing task", zap.Error(err))
				}
				w.bo.Reset()
			}
			if done := w.waitFor(ctx, w.bo.GetBackOffDuration()); done {
				break loop
			}
		}
	}
}

func (w *WorkerThread[T, R]) processTask(ctx context.Context, task *T) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\nstack: %s", r, string(debug.Stack()))
		}
	}()
	result, err := w.taskProcessor.ProcessTask(ctx, task)
	if err != nil {
		w.logger.Error("Error while processing task", zap.Error(err))
		err = w.taskProcessor.AbandonTask(ctx, task, ptr.Ptr(err.Error()))
		if err != nil {
			w.logger.Error("Error while abandoning task", zap.Error(err))
		}
	} else {
		w.logger.Debug("Complete task", zap.Any("result", result))
		err = w.taskProcessor.CompleteTask(ctx, result)
		if err != nil {
			w.logger.Error("Error while completing task", zap.Error(err))
			err = w.taskProcessor.AbandonTask(ctx, task, nil)
			if err != nil {
				w.logger.Error("Error while abandoning task", zap.Error(err))
			}
		}
	}
	return err
}

func (w *WorkerThread[T, R]) waitFor(ctx context.Context, duration time.Duration) bool {
	t := time.NewTimer(duration)
	select {
	case <-ctx.Done():
		if !t.Stop() {
			<-t.C
		}
		return true
	case <-t.C:
		return false
	}
}
