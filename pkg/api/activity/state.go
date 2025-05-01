package activity

import (
	"context"
	"github.com/tuannh982/simple-workflow-go/internal/activity"
)

func GetState(ctx context.Context) []byte {
	activityCtx := activity.MustExtractActivityExecutionContext(ctx)
	return activityCtx.GetStateData()
}

func UpdateState(ctx context.Context, update []byte) {
	activityCtx := activity.MustExtractActivityExecutionContext(ctx)
	activityCtx.UpdatedStateData = update
}
