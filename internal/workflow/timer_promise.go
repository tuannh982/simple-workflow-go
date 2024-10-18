package workflow

import "github.com/tuannh982/simple-workflows-go/internal/promise"

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
