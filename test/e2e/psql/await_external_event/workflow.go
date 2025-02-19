//go:build e2e
// +build e2e

package await_external_event

import (
	"context"
	"fmt"
	"github.com/tuannh982/simple-workflow-go/pkg/api/workflow"
)

type Void struct{}

type String struct {
	Value string
}

func Activity1(ctx context.Context, input *String) (*String, error) {
	return &String{
		Value: fmt.Sprintf("Hello, %s!", input.Value),
	}, nil
}

const HelloEventName = "hello"

func AwaitExternalEventWorkflow(ctx context.Context, _ *Void) (*String, error) {
	message, err := workflow.AwaitEvent(ctx, HelloEventName)
	if err != nil {
		return nil, err
	}
	activity1Result, err := workflow.CallActivity(ctx, Activity1, &String{Value: string(message)}).Await()
	if err != nil {
		return nil, err
	}
	return &String{Value: activity1Result.Value}, nil
}
