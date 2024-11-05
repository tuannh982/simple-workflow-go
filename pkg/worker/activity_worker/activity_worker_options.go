package activity_worker

import (
	"github.com/tuannh982/simple-workflows-go/internal/activity"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/worker"
	"time"
)

type ActivityWorkerOptions struct {
	WorkerOptions                *worker.WorkerOptions
	ActivityTaskProcessorOptions *activity.ActivityTaskProcessorOptions
}

func NewActivityWorkerOptions() *ActivityWorkerOptions {
	return &ActivityWorkerOptions{
		WorkerOptions:                worker.NewWorkerOptions(),
		ActivityTaskProcessorOptions: activity.NewActivityTaskProcessorOptions(),
	}
}

func WithMaxConcurrentTasksLimit(limit int) func(*ActivityWorkerOptions) {
	return func(options *ActivityWorkerOptions) {
		options.WorkerOptions.MaxConcurrentTasksLimit = limit
	}
}

func WithPollerInitialBackoffInterval(duration time.Duration) func(*ActivityWorkerOptions) {
	return func(options *ActivityWorkerOptions) {
		options.WorkerOptions.PollerInitialBackoffInterval = duration
	}
}

func WithPollerMaxBackoffInterval(duration time.Duration) func(*ActivityWorkerOptions) {
	return func(options *ActivityWorkerOptions) {
		options.WorkerOptions.PollerMaxBackoffInterval = duration
	}
}

func WithPollerBackoffMultiplier(multiplier float64) func(*ActivityWorkerOptions) {
	return func(options *ActivityWorkerOptions) {
		options.WorkerOptions.PollerBackoffMultiplier = multiplier
	}
}

func WithTaskProcessorInitialBackoffInterval(duration time.Duration) func(*ActivityWorkerOptions) {
	return func(options *ActivityWorkerOptions) {
		options.ActivityTaskProcessorOptions.InitialBackoffInterval = duration
	}
}

func WithTaskProcessorMaxBackoffInterval(duration time.Duration) func(*ActivityWorkerOptions) {
	return func(options *ActivityWorkerOptions) {
		options.ActivityTaskProcessorOptions.MaxBackoffInterval = duration
	}
}

func WithTaskProcessorBackoffMultiplier(multiplier float64) func(*ActivityWorkerOptions) {
	return func(options *ActivityWorkerOptions) {
		options.ActivityTaskProcessorOptions.BackoffMultiplier = multiplier
	}
}
