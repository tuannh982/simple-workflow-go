package workflow

import (
	"github.com/tuannh982/simple-workflow-go/pkg/utils/promise"
)

type TimerPromise struct {
	WorkflowRuntime *WorkflowRuntime
	Promise         *promise.Promise[struct{}]
}

func NewTimerPromise(runtime *WorkflowRuntime) *TimerPromise {
	p := promise.NewPromise[struct{}]()
	return &TimerPromise{
		WorkflowRuntime: runtime,
		Promise:         p,
	}
}

func (a *TimerPromise) Await() error {
	for {
		if a.Promise.IsDone() {
			if a.Promise.Error() != nil {
				return a.Promise.Error()
			} else {
				return nil
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
	panic(ErrControlledPanic)
}
