package api

import (
	"github.com/tuannh982/simple-workflows-go/internal/workflow"
)

type awaitableTimer struct {
	Promise *workflow.TimerPromise
}

func (t *awaitableTimer) Await() (any, error) {
	err := t.Promise.Await()
	return nil, err
}
