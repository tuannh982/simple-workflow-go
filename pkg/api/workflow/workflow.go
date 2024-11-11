package workflow

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/internal/workflow"
	"github.com/tuannh982/simple-workflows-go/pkg/types"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/awaitable"
	"time"
)

func GetWorkflowExecutionStartedTimestamp(ctx context.Context) int64 {
	workflowCtx := workflow.MustExtractWorkflowExecutionContext(ctx)
	return workflowCtx.WorkflowRuntime.WorkflowExecutionStartedTimestamp
}

func GetVersion(ctx context.Context) string {
	workflowCtx := workflow.MustExtractWorkflowExecutionContext(ctx)
	return workflowCtx.WorkflowRuntime.Version
}

func CallActivity[T any, R any](ctx context.Context, activity types.Activity[T, R], input *T) awaitable.Awaitable[*R] {
	workflowCtx := workflow.MustExtractWorkflowExecutionContext(ctx)
	promise := workflowCtx.WorkflowRuntime.ScheduleNewActivity(activity, input)
	return &AwaitableActivity[R]{
		Activity: activity,
		Promise:  promise,
	}
}

func GetCurrentTimestamp(ctx context.Context) time.Time {
	workflowCtx := workflow.MustExtractWorkflowExecutionContext(ctx)
	return time.UnixMilli(workflowCtx.WorkflowRuntime.CurrentTimestamp)
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
