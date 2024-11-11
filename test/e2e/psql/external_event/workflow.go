//go:build e2e
// +build e2e

package external_event

import (
	"context"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/pkg/api/workflow"
	"time"
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

func ExternalEventWorkflow(ctx context.Context, _ *Void) (*String, error) {
	message := ""
	workflow.OnEvent(ctx, HelloEventName, func(bytes []byte) {
		message = string(bytes)
	})
	workflow.WaitFor(ctx, 3*time.Second)
	activity1Result, err := workflow.CallActivity(ctx, Activity1, &String{Value: message}).Await()
	if err != nil {
		return nil, err
	}
	return &String{Value: activity1Result.Value}, nil
}
