package api

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/internal/workflow"
)

type Activity[T any, R any] func(context.Context, *T) (*R, error)

type AwaitableActivity[R any] struct {
	Activity any
	Promise  *workflow.ActivityPromise
}

func (t *AwaitableActivity[R]) Await() (*R, error) {
	ptr, err := t.Promise.Await()
	if ptr == nil {
		return nil, err
	} else {
		return ptr.(*R), err
	}
}
