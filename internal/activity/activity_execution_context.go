package activity

import "context"

type ActivityExecutionContext struct {
}

func NewActivityExecutionContext() *ActivityExecutionContext {
	return &ActivityExecutionContext{}
}

const ActivityExecutionContextKey = "activityExecutionContext"

func InjectActivityExecutionContext(ctx context.Context, activityExecutionContext *ActivityExecutionContext) context.Context {
	return context.WithValue(ctx, ActivityExecutionContextKey, activityExecutionContext)
}

func ExtractActivityExecutionContext(ctx context.Context) *ActivityExecutionContext {
	return ctx.Value(ActivityExecutionContextKey).(*ActivityExecutionContext)
}
