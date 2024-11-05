package worker

import (
	"context"
	"errors"
)

var ErrNoTask = errors.New("no task found")
var SuppressedError = errors.New("suppressed error")

type Summarizer interface {
	Summary() any
}

type TaskProcessor[T Summarizer, R Summarizer] interface {
	GetTask(ctx context.Context) (T, error)
	ProcessTask(ctx context.Context, task T) (R, error)
	CompleteTask(ctx context.Context, result R) error
	AbandonTask(ctx context.Context, task T, reason *string) error
}
