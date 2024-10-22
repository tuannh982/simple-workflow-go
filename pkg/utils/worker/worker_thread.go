package worker

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/backoff"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/ptr"
	"runtime/debug"
	"sync"
	"time"
)

type WorkerThread[T any, R any] struct {
	name          string
	taskProcessor TaskProcessor[T, R]
	bo            backoff.BackOff
	wg            *sync.WaitGroup
}

func NewWorkerThread[T any, R any](
	name string,
	taskProcessor TaskProcessor[T, R],
	bo backoff.BackOff,
	wg *sync.WaitGroup,
) *WorkerThread[T, R] {
	return &WorkerThread[T, R]{
		name:          name,
		taskProcessor: taskProcessor,
		bo:            bo,
		wg:            wg,
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
			log.Debug().
				Str("worker_thread", w.name).
				Msg("fetching task")
			task, err := w.taskProcessor.GetTask(ctx)
			log.Debug().
				Str("worker_thread", w.name).
				Msg("task fetched")
			if err != nil {
				if errors.Is(err, ErrNoTask) {
					// expected error, do not log
				} else {
					log.Err(err).
						Str("worker_thread", w.name).
						Msg("error while polling task")
				}
				w.bo.BackOff()
			} else {
				if err = w.processTask(ctx, task); err != nil {
					log.Err(err).
						Str("worker_thread", w.name).
						Msg("error while processing task")
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
		log.Err(err).
			Str("worker_thread", w.name).
			Msg("error while processing task")
		err = w.taskProcessor.AbandonTask(ctx, task, ptr.Ptr(err.Error()))
		if err != nil {
			log.Err(err).
				Str("worker_thread", w.name).
				Msg("error while abandoning task")
		}
	} else {
		err = w.taskProcessor.CompleteTask(ctx, result)
		if err != nil {
			log.Err(err).
				Str("worker_thread", w.name).
				Msg("error while completing task")
			err = w.taskProcessor.AbandonTask(ctx, task, nil)
			if err != nil {
				log.Err(err).
					Str("worker_thread", w.name).
					Msg("error while abandoning task")
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
