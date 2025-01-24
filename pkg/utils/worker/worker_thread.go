package worker

import (
	"context"
	"errors"
	"fmt"
	"github.com/tuannh982/simple-workflow-go/pkg/utils/backoff"
	"go.uber.org/zap"
	"runtime/debug"
	"sync"
	"time"
)

type WorkerThread interface {
	Run(ctx context.Context)
}

type workerThread[T Summarizer, R Summarizer] struct {
	name          string
	taskProcessor TaskProcessor[T, R]
	bo            backoff.Backoff
	wg            *sync.WaitGroup
	logger        *zap.Logger
}

func NewWorkerThread[T Summarizer, R Summarizer](
	name string,
	taskProcessor TaskProcessor[T, R],
	bo backoff.Backoff,
	wg *sync.WaitGroup,
	logger *zap.Logger,
) WorkerThread {
	return &workerThread[T, R]{
		name:          name,
		taskProcessor: taskProcessor,
		bo:            bo,
		wg:            wg,
		logger:        logger,
	}
}

func (w *workerThread[T, R]) Run(ctx context.Context) {
	w.logger.Info("WorkerThread started")
	defer func() {
		w.logger.Info("WorkerThread completed")
		w.wg.Done()
	}()
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			task, err := w.taskProcessor.GetTask(ctx)
			if err != nil {
				if !(errors.Is(err, ErrNoTask) || errors.Is(err, SuppressedError)) {
					w.logger.Error("Error while polling task", zap.Error(err))
				}
				w.bo.Backoff()
			} else {
				w.logger.Debug("Task fetched", zap.Any("task", task.Summary()))
				if err = w.processTask(ctx, task); err != nil {
					if !errors.Is(err, SuppressedError) {
						w.logger.Error("Error while processing task", zap.Error(err))
					}
				}
				w.bo.Reset()
			}
			if done := w.waitFor(ctx, w.bo.GetBackoffDuration()); done {
				break loop
			}
		}
	}
}

func (w *workerThread[T, R]) processTask(ctx context.Context, task T) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\nstack: %s", r, string(debug.Stack()))
		}
	}()
	result, err := w.taskProcessor.ProcessTask(ctx, task)
	if err != nil {
		w.logger.Error("Error while processing task", zap.Error(err), zap.Any("task", task.Summary()))
		err = w.taskProcessor.AbandonTask(ctx, result)
		if err != nil {
			if !errors.Is(err, SuppressedError) {
				w.logger.Error("Error while abandoning task", zap.Error(err), zap.Any("task", task.Summary()))
			}
		}
	} else {
		w.logger.Debug("Complete task", zap.Any("result", result.Summary()))
		err = w.taskProcessor.CompleteTask(ctx, result)
		if err != nil {
			if !errors.Is(err, SuppressedError) {
				w.logger.Error("Error while completing task", zap.Error(err), zap.Any("task", task.Summary()))
			}
			err = w.taskProcessor.AbandonTask(ctx, result)
			if err != nil {
				if !errors.Is(err, SuppressedError) {
					w.logger.Error("Error while abandoning task", zap.Error(err), zap.Any("task", task.Summary()))
				}
			}
		}
	}
	return err
}

func (w *workerThread[T, R]) waitFor(ctx context.Context, duration time.Duration) bool {
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
