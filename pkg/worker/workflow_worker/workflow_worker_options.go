package workflow_worker

import (
	"github.com/tuannh982/simple-workflows-go/pkg/utils/worker"
	"time"
)

type WorkflowWorkerOptions struct {
	WorkerOptions *worker.WorkerOptions
}

func NewWorkflowWorkerOptions() *WorkflowWorkerOptions {
	return &WorkflowWorkerOptions{
		WorkerOptions: worker.NewWorkerOptions(),
	}
}

func WithMaxConcurrentTasksLimit(limit int) func(*WorkflowWorkerOptions) {
	return func(options *WorkflowWorkerOptions) {
		options.WorkerOptions.MaxConcurrentTasksLimit = limit
	}
}

func WithPollerInitialBackoffInterval(duration time.Duration) func(*WorkflowWorkerOptions) {
	return func(options *WorkflowWorkerOptions) {
		options.WorkerOptions.PollerInitialBackoffInterval = duration
	}
}

func WithPollerMaxBackoffInterval(duration time.Duration) func(*WorkflowWorkerOptions) {
	return func(options *WorkflowWorkerOptions) {
		options.WorkerOptions.PollerMaxBackoffInterval = duration
	}
}
