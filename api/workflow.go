package api

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/internal/workflow"
	"time"
)

func StartActivity[T any, R any](ctx context.Context, activity Activity[T, R], input *T) Awaitable[*R] {
	workflowCtx := workflow.ExtractWorkflowExecutionContext(ctx)
	promise := workflowCtx.WorkflowRuntime.ScheduleNewActivity(activity, input)
	return &awaitableActivity[R]{
		Activity: activity,
		Promise:  promise,
	}
}

func CreateTimer(ctx context.Context, delay time.Duration) Awaitable[any] {
	workflowCtx := workflow.ExtractWorkflowExecutionContext(ctx)
	fireAtTimestamp := workflowCtx.WorkflowRuntime.CurrentTimestamp + delay.Milliseconds()
	promise := workflowCtx.WorkflowRuntime.CreateTimer(fireAtTimestamp)
	return &awaitableTimer{
		Promise: promise,
	}
}

func CreateTimerAt(ctx context.Context, fireAt time.Time) Awaitable[any] {
	workflowCtx := workflow.ExtractWorkflowExecutionContext(ctx)
	fireAtTimestamp := fireAt.UnixMilli()
	promise := workflowCtx.WorkflowRuntime.CreateTimer(fireAtTimestamp)
	return &awaitableTimer{
		Promise: promise,
	}
}
