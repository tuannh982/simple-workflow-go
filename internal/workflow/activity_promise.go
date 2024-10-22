package workflow

import (
	"github.com/tuannh982/simple-workflows-go/internal/fn"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/promise"
)

type ActivityPromise struct {
	WorkflowRuntime *WorkflowRuntime
	Activity        any
	Promise         *promise.Promise[[]byte]
}

func NewActivityPromise(runtime *WorkflowRuntime, activity any) *ActivityPromise {
	p := promise.NewPromise[[]byte]()
	return &ActivityPromise{
		WorkflowRuntime: runtime,
		Activity:        activity,
		Promise:         p,
	}
}

func (a *ActivityPromise) Await() (any, error) {
	for {
		if a.Promise.IsDone() {
			err := a.Promise.Error()
			if a.Promise.Value() != nil {
				ptr := fn.InitResult(a.Activity)
				bytes := *a.Promise.Value()
				unExpectedErr := a.WorkflowRuntime.DataConverter.Unmarshal(bytes, ptr)
				if unExpectedErr != nil {
					panic(unExpectedErr)
				}
				return ptr, err
			} else {
				return nil, err
			}
		} else {
			done, unExpectedErr := a.WorkflowRuntime.Step()
			if unExpectedErr != nil {
				panic(unExpectedErr)
			}
			if done {
				break
			}
		}
	}
	if a.Promise.IsDone() {
		err := a.Promise.Error()
		if a.Promise.Value() != nil {
			ptr := fn.InitResult(a.Activity)
			bytes := *a.Promise.Value()
			unExpectedErr := a.WorkflowRuntime.DataConverter.Unmarshal(bytes, ptr)
			if unExpectedErr != nil {
				panic(unExpectedErr)
			}
			return ptr, err
		} else {
			return nil, err
		}
	} else {
		panic(ErrControlledPanic)
	}
}
