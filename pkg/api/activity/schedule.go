package activity

import (
	"context"
	"github.com/tuannh982/simple-workflow-go/internal/activity"
	"time"
)

func OverrideNextExecutionTime(ctx context.Context, t time.Time) {
	activityCtx := activity.MustExtractActivityExecutionContext(ctx)
	activityCtx.UserDefinedNextExecutionTime = &t
}

func OverrideBackoffDuration(ctx context.Context, delay time.Duration) {
	activityCtx := activity.MustExtractActivityExecutionContext(ctx)
	activityCtx.UserDefinedBackoffDuration = &delay
}
