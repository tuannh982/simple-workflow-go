package task

import (
	"github.com/tuannh982/simple-workflows-go/pkg/dto"
)

type ActivityTaskSummary struct {
	TaskID          string
	WorkflowID      string
	TaskScheduledID int64
	NumAttempted    int
}

func (t *ActivityTask) Summary() any {
	return &ActivityTaskSummary{
		TaskID:          t.TaskID,
		WorkflowID:      t.WorkflowID,
		TaskScheduledID: t.TaskScheduleEvent.TaskScheduledID,
		NumAttempted:    t.NumAttempted,
	}
}

type ActivityTaskResultSummary struct {
	TaskID          string
	WorkflowID      string
	TaskScheduledID int64
	ExecutionResult *dto.ExecutionResult
}

func (r *ActivityTaskResult) Summary() any {
	return &ActivityTaskResultSummary{
		TaskID:          r.Task.TaskID,
		WorkflowID:      r.Task.WorkflowID,
		TaskScheduledID: r.Task.TaskScheduleEvent.TaskScheduledID,
		ExecutionResult: r.ExecutionResult,
	}
}
