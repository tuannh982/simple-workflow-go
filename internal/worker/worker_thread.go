package worker

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/tuannh982/simple-workflows-go/internal/backoff"
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
	for {
		select {
		case <-ctx.Done():
			return
		default:
			log.Debug().
				Str("worker_thread", w.name).
				Msg("fetching task")
			task, err := w.taskProcessor.GetTask(ctx)
			log.Debug().
				Str("worker_thread", w.name).
				Msg("task fetched")
			if err != nil {
				log.Err(err).
					Str("worker_thread", w.name).
					Msg("error while polling task")
				if errors.Is(err, ErrNoTask) {
					log.Err(err).
						Str("worker_thread", w.name).
						Msg("no task found")
				}
				w.bo.BackOff()
			} else {
				err = w.processTask(ctx, task)
				if err == nil {
					w.bo.Reset()
				}
			}
			t := time.NewTimer(w.bo.GetBackOffInterval())
			select {
			case <-ctx.Done():
				if !t.Stop() {
					<-t.C
				}
				return
			case <-t.C:
			}
		}
	}
}

func (w *WorkerThread[T, R]) processTask(ctx context.Context, task *T) error {
	var err error
	result, err := w.taskProcessor.ProcessTask(ctx, task)
	if err != nil {
		log.Err(err).
			Str("worker_thread", w.name).
			Msg("error while processing task")
		err = w.taskProcessor.AbandonTask(ctx, task)
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
			err = w.taskProcessor.AbandonTask(ctx, task)
			if err != nil {
				log.Err(err).
					Str("worker_thread", w.name).
					Msg("error while abandoning task")
			}
		}
	}
	return err
}
