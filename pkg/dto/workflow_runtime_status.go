package dto

type WorkflowRuntimeStatus string

const (
	WorkflowRuntimeStatusPending   WorkflowRuntimeStatus = "pending"
	WorkflowRuntimeStatusRunning   WorkflowRuntimeStatus = "running"
	WorkflowRuntimeStatusCompleted WorkflowRuntimeStatus = "completed"
)
