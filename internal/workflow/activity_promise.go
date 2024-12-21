package workflow

import (
	"github.com/tuannh982/simple-workflow-go/internal/fn"
	"github.com/tuannh982/simple-workflow-go/pkg/utils/promise"
)

type UntypedActivityPromise struct {
	WorkflowRuntime *WorkflowRuntime
	Promise         *promise.Promise[[]byte]
}

func NewUntypedActivityPromise(runtime *WorkflowRuntime) *UntypedActivityPromise {
	p := promise.NewPromise[[]byte]()
	return &UntypedActivityPromise{
		WorkflowRuntime: runtime,
		Promise:         p,
	}
}

func (a *UntypedActivityPromise) ToTyped(activity any) *ActivityPromise {
	return &ActivityPromise{
		WorkflowRuntime: a.WorkflowRuntime,
		Activity:        activity,
		Promise:         a.Promise,
	}
}

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
