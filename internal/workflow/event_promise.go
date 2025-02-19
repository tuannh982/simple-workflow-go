package workflow

import "github.com/tuannh982/simple-workflow-go/pkg/utils/promise"

type EventPromise struct {
	WorkflowRuntime *WorkflowRuntime
	EventName       string
	Promise         *promise.Promise[[]byte]
}

func NewEventPromise(runtime *WorkflowRuntime, eventName string) *EventPromise {
	p := promise.NewPromise[[]byte]()
	return &EventPromise{
		WorkflowRuntime: runtime,
		EventName:       eventName,
		Promise:         p,
	}
}

func (a *EventPromise) Await() ([]byte, error) {
	for {
		if a.Promise.IsDone() {
			err := a.Promise.Error()
			if a.Promise.Value() != nil {
				return *a.Promise.Value(), nil
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
	panic(ErrControlledPanic)
}
