package workflow

import (
	"context"
	"github.com/tuannh982/simple-workflow-go/internal/workflow"
	"github.com/tuannh982/simple-workflow-go/pkg/types"
	"github.com/tuannh982/simple-workflow-go/pkg/utils/awaitable"
	"time"
)

func CallActivity[T any, R any](ctx context.Context, activity types.Activity[T, R], input *T) awaitable.Awaitable[*R] {
	workflowCtx := workflow.MustExtractWorkflowExecutionContext(ctx)
	promise := workflowCtx.WorkflowRuntime.ScheduleNewActivity(activity, input)
	return &AwaitableActivity[R]{
		Activity: activity,
		Promise:  promise,
	}
}

func WaitFor(ctx context.Context, delay time.Duration) {
	workflowCtx := workflow.MustExtractWorkflowExecutionContext(ctx)
	fireAtTimestamp := workflowCtx.WorkflowRuntime.CurrentTimestamp + delay.Milliseconds()
	promise := workflowCtx.WorkflowRuntime.CreateTimer(fireAtTimestamp)
	a := &AwaitableTimer{
		Promise: promise,
	}
	_, _ = a.Await()
}

func WaitUntil(ctx context.Context, until time.Time) {
	workflowCtx := workflow.MustExtractWorkflowExecutionContext(ctx)
	fireAtTimestamp := until.UnixMilli()
	promise := workflowCtx.WorkflowRuntime.CreateTimer(fireAtTimestamp)
	a := &AwaitableTimer{
		Promise: promise,
	}
	_, _ = a.Await()
}

func SetVar[T any](ctx context.Context, name string, value T) {
	workflowCtx := workflow.MustExtractWorkflowExecutionContext(ctx)
	workflowCtx.UserDefinedVars[name] = value
}

func OnEvent(ctx context.Context, eventName string, callback func([]byte)) {
	workflowCtx := workflow.MustExtractWorkflowExecutionContext(ctx)
	if _, ok := workflowCtx.EventCallbacks[eventName]; !ok {
		workflowCtx.EventCallbacks[eventName] = make([]func([]byte), 0)
	}
	workflowCtx.EventCallbacks[eventName] = append(workflowCtx.EventCallbacks[eventName], callback)
}

func AwaitEvent(ctx context.Context, eventName string) ([]byte, error) {
	workflowCtx := workflow.MustExtractWorkflowExecutionContext(ctx)
	promise := workflowCtx.WorkflowRuntime.NewEventPromise(eventName)
	return promise.Await()
}
