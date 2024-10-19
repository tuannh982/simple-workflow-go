package api

import (
	"github.com/tuannh982/simple-workflows-go/internal/workflow"
)

type AwaitableTimer struct {
	Promise *workflow.TimerPromise
}

func (t *AwaitableTimer) Await() (any, error) {
	err := t.Promise.Await()
	return nil, err
}
