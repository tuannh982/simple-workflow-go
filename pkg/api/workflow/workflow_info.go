package workflow

import (
	"context"
	"github.com/tuannh982/simple-workflow-go/internal/workflow"
	"time"
)

type WorkflowInfo struct {
	Version                           string
	WorkflowExecutionStartedTimestamp time.Time
	CurrentTimestamp                  time.Time
	IsReplaying                       bool
}

func GetWorkflowInfo(ctx context.Context) *WorkflowInfo {
	workflowCtx := workflow.MustExtractWorkflowExecutionContext(ctx)
	return &WorkflowInfo{
		Version:                           workflowCtx.WorkflowRuntime.Version,
		WorkflowExecutionStartedTimestamp: time.UnixMilli(workflowCtx.WorkflowRuntime.WorkflowExecutionStartedTimestamp),
		CurrentTimestamp:                  time.UnixMilli(workflowCtx.WorkflowRuntime.CurrentTimestamp),
		IsReplaying:                       workflowCtx.WorkflowRuntime.IsReplaying,
	}
}

func GetVersion(ctx context.Context) string {
	return GetWorkflowInfo(ctx).Version
}

func GetWorkflowExecutionStartedTimestamp(ctx context.Context) time.Time {
	return GetWorkflowInfo(ctx).WorkflowExecutionStartedTimestamp
}

func GetCurrentTimestamp(ctx context.Context) time.Time {
	return GetWorkflowInfo(ctx).CurrentTimestamp
}

func IsReplaying(ctx context.Context) bool {
	return GetWorkflowInfo(ctx).IsReplaying
}
