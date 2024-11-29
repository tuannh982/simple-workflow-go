package workflow

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
