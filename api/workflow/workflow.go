package workflow

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/internal/activity"
	"github.com/tuannh982/simple-workflows-go/internal/workflow"
	"time"
)

func StartActivity(ctx context.Context, activity activity.Activity, input any) AwaitableActivity {
	workflowCtx := workflow.ExtractWorkflowExecutionContext(ctx)
	awaitable := workflowCtx.WorkflowRuntime.ScheduleNewActivity(activity, input)
	return awaitable
}

func CreateTimer(ctx context.Context, delay time.Duration) AwaitableTimer {
	workflowCtx := workflow.ExtractWorkflowExecutionContext(ctx)
	fireAtTimestamp := workflowCtx.WorkflowRuntime.CurrentTimestamp + delay.Milliseconds()
	awaitable := workflowCtx.WorkflowRuntime.CreateTimer(fireAtTimestamp)
	return awaitable
}

func CreateTimerAt(ctx context.Context, fireAt time.Time) AwaitableTimer {
	workflowCtx := workflow.ExtractWorkflowExecutionContext(ctx)
	fireAtTimestamp := fireAt.UnixMilli()
	awaitable := workflowCtx.WorkflowRuntime.CreateTimer(fireAtTimestamp)
	return awaitable
}
