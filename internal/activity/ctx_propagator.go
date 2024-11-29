package activity

import "context"

const ActivityExecutionContextKey = "activityExecutionContext"

func InjectActivityExecutionContext(ctx context.Context, activityExecutionContext *ActivityExecutionContext) context.Context {
	return context.WithValue(ctx, ActivityExecutionContextKey, activityExecutionContext)
}

func MustExtractActivityExecutionContext(ctx context.Context) *ActivityExecutionContext {
	return ctx.Value(ActivityExecutionContextKey).(*ActivityExecutionContext)
}
