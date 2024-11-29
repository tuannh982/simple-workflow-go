package workflow

import "context"

const WorkflowExecutionContextKey = "workflowExecutionContext"

func InjectWorkflowExecutionContext(ctx context.Context, workflowExecutionContext *WorkflowExecutionContext) context.Context {
	return context.WithValue(ctx, WorkflowExecutionContextKey, workflowExecutionContext)
}

func MustExtractWorkflowExecutionContext(ctx context.Context) *WorkflowExecutionContext {
	return ctx.Value(WorkflowExecutionContextKey).(*WorkflowExecutionContext)
}
