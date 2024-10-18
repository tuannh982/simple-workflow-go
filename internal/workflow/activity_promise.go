package workflow

import (
	"github.com/tuannh982/simple-workflows-go/internal/promise"
)

type AwaitableActivity interface {
	Await(vPtr any) error
}

type ActivityPromise struct {
	WorkflowRuntime *WorkflowRuntime
	Promise         *promise.Promise[[]byte]
}

func NewActivityPromise(runtime *WorkflowRuntime) *ActivityPromise {
	p := promise.NewPromise[[]byte]()
	return &ActivityPromise{
		WorkflowRuntime: runtime,
		Promise:         p,
	}
}

func (a *ActivityPromise) Await(vPtr any) error {
	for {
		if a.Promise.IsDone() {
			if a.Promise.Error() != nil {
				return a.Promise.Error()
			} else {
				var bytes []byte
				if a.Promise.Value() != nil {
					bytes = *a.Promise.Value()
				} else {
					bytes = []byte{}
				}
				return a.WorkflowRuntime.DataConverter.Unmarshal(bytes, vPtr)
			}
		} else {
			done, unexpectedErr := a.WorkflowRuntime.Step()
			if unexpectedErr != nil {
				panic(unexpectedErr)
			}
			if done {
				break
			}
		}
	}
	panic(ErrControlledPanic)
}
