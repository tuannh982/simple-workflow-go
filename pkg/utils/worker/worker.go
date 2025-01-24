package worker

import (
	"context"
	"fmt"
	"github.com/tuannh982/simple-workflow-go/pkg/utils/backoff"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Worker interface {
	Start(ctx context.Context)
	Stop(ctx context.Context)
}

type worker[T Summarizer, R Summarizer] struct {
	name                         string
	taskProcessor                TaskProcessor[T, R]
	logger                       *zap.Logger
	cancel                       context.CancelFunc
	wg                           *sync.WaitGroup
	maxConcurrentTasks           int
	pollerInitialBackoffInterval time.Duration
	pollerMaxBackoffInterval     time.Duration
	pollerBackoffMultiplier      float64
}

func NewWorkerOpts[T Summarizer, R Summarizer](
	name string,
	taskProcessor TaskProcessor[T, R],
	logger *zap.Logger,
	opts ...func(*WorkerOptions),
) Worker {
	options := NewWorkerOptions()
	for _, configure := range opts {
		configure(options)
	}
	return NewWorker(name, taskProcessor, logger, options)
}

func NewWorker[T Summarizer, R Summarizer](
	name string,
	taskProcessor TaskProcessor[T, R],
	logger *zap.Logger,
	options *WorkerOptions,
) Worker {
	return &worker[T, R]{
		name:                         name,
		taskProcessor:                taskProcessor,
		logger:                       logger,
		wg:                           &sync.WaitGroup{},
		maxConcurrentTasks:           options.MaxConcurrentTasksLimit,
		pollerInitialBackoffInterval: options.PollerInitialBackoffInterval,
		pollerMaxBackoffInterval:     options.PollerMaxBackoffInterval,
		pollerBackoffMultiplier:      options.PollerBackoffMultiplier,
	}
}

func (w *worker[T, R]) Start(ctx context.Context) {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	for i := 0; i < w.maxConcurrentTasks; i++ {
		threadName := fmt.Sprintf("Thread %d", i)
		thread := NewWorkerThread(
			threadName,
			w.taskProcessor,
			backoff.NewExponentialBackoff(
				w.pollerInitialBackoffInterval,
				w.pollerMaxBackoffInterval,
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
