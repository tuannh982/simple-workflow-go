package workflow

import (
	"github.com/tuannh982/simple-workflow-go/internal/workflow"
)

type AwaitableTimer struct {
	Promise *workflow.TimerPromise
}

func (t *AwaitableTimer) Await() (any, error) {
	err := t.Promise.Await()
	return nil, err
}
