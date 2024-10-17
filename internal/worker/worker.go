package worker

import (
	"context"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/internal/backoff"
	"sync"
	"time"
)

type Worker[T any, R any] interface {
	Start(ctx context.Context)
	Stop(ctx context.Context)
}

type worker[T any, R any] struct {
	name                    string
	taskProcessor           TaskProcessor[T, R]
	cancel                  context.CancelFunc
	wg                      *sync.WaitGroup
	maxConcurrentTasks      int
	pollerInitialInterval   time.Duration
	pollerMaxInterval       time.Duration
	pollerBackoffMultiplier float64
}

func NewWorker[T any, R any](
	name string,
	taskProcessor TaskProcessor[T, R],
	opts ...func(*WorkerOptions),
) Worker[T, R] {
	options := newWorkerOptions()
	for _, configure := range opts {
		configure(options)
	}
	return &worker[T, R]{
		name:                    name,
		taskProcessor:           taskProcessor,
		wg:                      &sync.WaitGroup{},
		maxConcurrentTasks:      options.maxConcurrentTasksLimit,
		pollerInitialInterval:   options.pollerInitialInterval,
		pollerMaxInterval:       options.pollerMaxInterval,
		pollerBackoffMultiplier: options.pollerBackoffMultiplier,
	}
}

func (w *worker[T, R]) Start(ctx context.Context) {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	for i := 0; i < w.maxConcurrentTasks; i++ {
		thread := NewWorkerThread[T, R](
			fmt.Sprintf("%s_thread_%d", w.name, i),
			w.taskProcessor,
			backoff.NewExponentialBackOff(
				w.pollerInitialInterval,
				w.pollerMaxInterval,
				w.pollerBackoffMultiplier,
			),
			w.wg,
		)
		w.wg.Add(1)
		go thread.Run(ctxWithCancel)
	}
}

func (w *worker[T, R]) Stop(ctx context.Context) {
	w.cancel()
	w.wg.Wait()
}
