package workflow

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/internal/workflow"
	"github.com/tuannh982/simple-workflows-go/pkg/types"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/awaitable"
	"time"
)

func GetVersion(ctx context.Context) string {
	workflowCtx := workflow.ExtractWorkflowExecutionContext(ctx)
	return workflowCtx.WorkflowRuntime.Version
}

func CallActivity[T any, R any](ctx context.Context, activity types.Activity[T, R], input *T) awaitable.Awaitable[*R] {
	workflowCtx := workflow.ExtractWorkflowExecutionContext(ctx)
	promise := workflowCtx.WorkflowRuntime.ScheduleNewActivity(activity, input)
	return &AwaitableActivity[R]{
		Activity: activity,
		Promise:  promise,
	}
}

func WaitFor(ctx context.Context, delay time.Duration) {
	workflowCtx := workflow.ExtractWorkflowExecutionContext(ctx)
	fireAtTimestamp := workflowCtx.WorkflowRuntime.CurrentTimestamp + delay.Milliseconds()
	promise := workflowCtx.WorkflowRuntime.CreateTimer(fireAtTimestamp)
	a := &AwaitableTimer{
		Promise: promise,
	}
	_, _ = a.Await()
}
