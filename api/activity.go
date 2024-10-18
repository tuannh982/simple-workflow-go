package api

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
	"github.com/tuannh982/simple-workflows-go/internal/workflow"
)

type Activity[T any, R any] func(context.Context, *T) (*R, error)

type awaitableActivity[R any] struct {
	Activity any
	Promise  *workflow.ActivityPromise
}

func (t *awaitableActivity[R]) Await() (*R, error) {
	ptr := fn.InitResult(t.Activity)
	err := t.Promise.Await(ptr)
	return ptr.(*R), err
}
