package worker

import "time"

type WorkerOptions struct {
	MaxConcurrentTasksLimit      int
	PollerInitialBackoffInterval time.Duration
	PollerMaxBackoffInterval     time.Duration
	PollerBackoffMultiplier      float64
}

func NewWorkerOptions() *WorkerOptions {
	return &WorkerOptions{
		MaxConcurrentTasksLimit:      1,
		PollerInitialBackoffInterval: 500 * time.Millisecond,
		PollerMaxBackoffInterval:     500 * time.Millisecond,
		PollerBackoffMultiplier:      1,
	}
}

func WithMaxConcurrentTasksLimit(limit int) func(*WorkerOptions) {
	return func(options *WorkerOptions) {
		options.MaxConcurrentTasksLimit = limit
	}
}

func WithPollerInitialBackoffInterval(duration time.Duration) func(*WorkerOptions) {
	return func(options *WorkerOptions) {
		options.PollerInitialBackoffInterval = duration
	}
}

func WithPollerMaxBackoffInterval(duration time.Duration) func(*WorkerOptions) {
	return func(options *WorkerOptions) {
		options.PollerMaxBackoffInterval = duration
	}
}

func WithPollerBackoffMultiplier(multiplier float64) func(*WorkerOptions) {
	return func(options *WorkerOptions) {
		options.PollerBackoffMultiplier = multiplier
	}
}
