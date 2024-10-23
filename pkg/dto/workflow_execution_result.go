package dto

type WorkflowExecutionResult struct {
	WorkflowID    string
	Version       string
	RuntimeStatus string
	ExecutionResult
}
