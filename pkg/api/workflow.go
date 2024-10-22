package api

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
	"github.com/tuannh982/simple-workflows-go/internal/workflow"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/types"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/awaitable"
	"time"
)

func CallActivity[T any, R any](ctx context.Context, activity types.Activity[T, R], input *T) awaitable.Awaitable[*R] {
	workflowCtx := workflow.ExtractWorkflowExecutionContext(ctx)
	promise := workflowCtx.WorkflowRuntime.ScheduleNewActivity(activity, input)
	return &AwaitableActivity[R]{
		Activity: activity,
		Promise:  promise,
	}
}

func CreateTimer(ctx context.Context, delay time.Duration) awaitable.Awaitable[any] {
	workflowCtx := workflow.ExtractWorkflowExecutionContext(ctx)
	fireAtTimestamp := workflowCtx.WorkflowRuntime.CurrentTimestamp + delay.Milliseconds()
	promise := workflowCtx.WorkflowRuntime.CreateTimer(fireAtTimestamp)
	return &AwaitableTimer{
		Promise: promise,
	}
}

func CreateTimerAt(ctx context.Context, fireAt time.Time) awaitable.Awaitable[any] {
	workflowCtx := workflow.ExtractWorkflowExecutionContext(ctx)
	fireAtTimestamp := fireAt.UnixMilli()
	promise := workflowCtx.WorkflowRuntime.CreateTimer(fireAtTimestamp)
	return &AwaitableTimer{
		Promise: promise,
	}
}

// TODO unstable api
func ScheduleWorkflow[T any, R any](
	dataConverter dataconverter.DataConverter,
	backend backend.Backend,
	workflow types.Workflow[T, R],
	input *T,
) error {
	name := fn.GetFunctionName(workflow)
	inputBytes, err := dataConverter.Marshal(input)
	if err != nil {
		panic(err)
	}
	executionStarted := &history.WorkflowExecutionStarted{
		Name:               name,
		Version:            "",
		Input:              inputBytes,
		WorkflowID:         "test",
		ParentWorkflowInfo: nil,
	}
	return backend.CreateWorkflow(context.TODO(), executionStarted)
}
