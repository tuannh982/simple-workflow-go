package worker

import (
	"context"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/backoff"
	"go.uber.org/zap"
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
	logger                  *zap.Logger
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
	logger *zap.Logger,
	opts ...func(*WorkerOptions),
) Worker[T, R] {
	options := newWorkerOptions()
	for _, configure := range opts {
		configure(options)
	}
	return &worker[T, R]{
		name:                    name,
		taskProcessor:           taskProcessor,
		logger:                  logger,
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
		threadName := fmt.Sprintf("Thread %d", i)
		thread := NewWorkerThread[T, R](
			threadName,
			w.taskProcessor,
			backoff.NewExponentialBackOff(
				w.pollerInitialInterval,
				w.pollerMaxInterval,
				w.pollerBackoffMultiplier,
			),
			w.wg,
			w.logger.With(zap.String("thread", threadName)),
		)
		w.wg.Add(1)
		go thread.Run(ctxWithCancel)
	}
}

func (w *worker[T, R]) Stop(_ context.Context) {
	w.cancel()
	w.wg.Wait()
}
