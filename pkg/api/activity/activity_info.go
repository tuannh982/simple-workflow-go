package activity

import (
	"context"
	"github.com/tuannh982/simple-workflow-go/internal/activity"
)

type ActivityInfo struct {
	WorkflowID   string
	NumAttempted int
}

func GetActivityInfo(ctx context.Context) *ActivityInfo {
	activityCtx := activity.MustExtractActivityExecutionContext(ctx)
	return &ActivityInfo{
		WorkflowID:   activityCtx.Task.WorkflowID,
		NumAttempted: activityCtx.Task.NumAttempted,
	}
}

func GetWorkflowID(ctx context.Context) string {
	return GetActivityInfo(ctx).WorkflowID
}

func GetNumAttempted(ctx context.Context) int {
	return GetActivityInfo(ctx).NumAttempted
}
