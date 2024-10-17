package history

import (
	"github.com/tuannh982/simple-workflows-go/internal/dto"
)

type ParentWorkflowInfo struct {
	TaskScheduledID int32
	Name            int32
	Version         int32
	WorkflowID      string
}

type WorkflowExecutionStarted struct {
	Name                     string
	Version                  string
	Input                    []byte
	WorkflowID               string
	ParentWorkflowInfo       *ParentWorkflowInfo
	ScheduleToStartTimestamp int64
}

type WorkflowExecutionCompleted struct {
	Status string
	dto.ExecutionResult
}

type WorkflowExecutionTerminated struct {
	Reason string
}
