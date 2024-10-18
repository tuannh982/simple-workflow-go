package workflow

import "context"

type WorkflowExecutionContext struct {
	WorkflowRuntime *WorkflowRuntime
}

func NewWorkflowExecutionContext(runtime *WorkflowRuntime) *WorkflowExecutionContext {
	return &WorkflowExecutionContext{
		WorkflowRuntime: runtime,
	}
}

const WorkflowExecutionContextKey = "workflowExecutionContext"

func InjectWorkflowExecutionContext(ctx context.Context, workflowExecutionContext *WorkflowExecutionContext) context.Context {
	return context.WithValue(ctx, WorkflowExecutionContextKey, workflowExecutionContext)
}

func ExtractWorkflowExecutionContext(ctx context.Context) *WorkflowExecutionContext {
	return ctx.Value(WorkflowExecutionContextKey).(*WorkflowExecutionContext)
}
