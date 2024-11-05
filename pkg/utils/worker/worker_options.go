package worker

import "time"

type WorkerOptions struct {
	maxConcurrentTasksLimit int
	pollerInitialInterval   time.Duration
	pollerMaxInterval       time.Duration
	pollerBackoffMultiplier float64
}

func newWorkerOptions() *WorkerOptions {
	return &WorkerOptions{
		maxConcurrentTasksLimit: 1,
		pollerInitialInterval:   500 * time.Millisecond,
		pollerMaxInterval:       500 * time.Millisecond,
		pollerBackoffMultiplier: 1,
	}
}

func WithMaxConcurrentTasksLimit(limit int) func(*WorkerOptions) {
	return func(options *WorkerOptions) {
		options.maxConcurrentTasksLimit = limit
	}
}

func WithPollerInitialInterval(duration time.Duration) func(*WorkerOptions) {
	return func(options *WorkerOptions) {
		options.pollerInitialInterval = duration
	}
}

func WithPollerMaxInterval(duration time.Duration) func(*WorkerOptions) {
	return func(options *WorkerOptions) {
		options.pollerMaxInterval = duration
	}
}

func WithPollerBackoffMultiplier(multiplier float64) func(*WorkerOptions) {
	return func(options *WorkerOptions) {
		options.pollerBackoffMultiplier = multiplier
	}
}
