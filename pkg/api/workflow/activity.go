package workflow

import (
	"github.com/tuannh982/simple-workflow-go/internal/workflow"
)

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
