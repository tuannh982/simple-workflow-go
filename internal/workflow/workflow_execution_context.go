package workflow

import (
	"context"
)

type WorkflowExecutionContext struct {
	WorkflowRuntime *WorkflowRuntime
	UserDefinedVars map[string]any
}

func NewWorkflowExecutionContext(runtime *WorkflowRuntime) *WorkflowExecutionContext {
	return &WorkflowExecutionContext{
		WorkflowRuntime: runtime,
		UserDefinedVars: make(map[string]any),
	}
}

const WorkflowExecutionContextKey = "workflowExecutionContext"

func InjectWorkflowExecutionContext(ctx context.Context, workflowExecutionContext *WorkflowExecutionContext) context.Context {
	return context.WithValue(ctx, WorkflowExecutionContextKey, workflowExecutionContext)
}

func MustExtractWorkflowExecutionContext(ctx context.Context) *WorkflowExecutionContext {
	return ctx.Value(WorkflowExecutionContextKey).(*WorkflowExecutionContext)
}
