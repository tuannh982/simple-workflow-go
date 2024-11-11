package workflow

import (
	"context"
)

type WorkflowExecutionContext struct {
	WorkflowRuntime *WorkflowRuntime
	UserDefinedVars map[string]any
	EventCallbacks  map[string][]func([]byte)
}

func NewWorkflowExecutionContext(runtime *WorkflowRuntime) *WorkflowExecutionContext {
	return &WorkflowExecutionContext{
		WorkflowRuntime: runtime,
		UserDefinedVars: make(map[string]any),
		EventCallbacks:  make(map[string][]func([]byte)),
	}
}

const WorkflowExecutionContextKey = "workflowExecutionContext"

func InjectWorkflowExecutionContext(ctx context.Context, workflowExecutionContext *WorkflowExecutionContext) context.Context {
	return context.WithValue(ctx, WorkflowExecutionContextKey, workflowExecutionContext)
}

func MustExtractWorkflowExecutionContext(ctx context.Context) *WorkflowExecutionContext {
	return ctx.Value(WorkflowExecutionContextKey).(*WorkflowExecutionContext)
}
