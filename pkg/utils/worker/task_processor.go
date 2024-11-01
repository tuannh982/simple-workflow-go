package worker

import (
	"context"
	"errors"
)

var ErrNoTask = errors.New("no task found")
var SuppressedError = errors.New("suppressed error")

type TaskProcessor[T any, R any] interface {
	GetTask(ctx context.Context) (*T, error)
	ProcessTask(ctx context.Context, task *T) (*R, error)
	CompleteTask(ctx context.Context, result *R) error
	AbandonTask(ctx context.Context, task *T, reason *string) error
}
