package task

import (
	"github.com/tuannh982/simple-workflow-go/pkg/dto"
	"github.com/tuannh982/simple-workflow-go/pkg/dto/history"
	"time"
)

type ActivityTask struct {
	TaskID            string
	WorkflowID        string
	NumAttempted      int
	StateData         []byte
	TaskScheduleEvent *history.ActivityScheduled
}

type ActivityTaskExecutionError struct {
	Error             error
	NextExecutionTime *time.Time
}

type ActivityTaskResult struct {
	Task             *ActivityTask
	UpdatedStateData []byte
	ExecutionResult  *dto.ExecutionResult
	ExecutionError   *ActivityTaskExecutionError
}

func (r *ActivityTaskResult) GetStateData() []byte {
	if r.UpdatedStateData != nil {
		return r.UpdatedStateData
	} else {
		return r.Task.StateData
	}
}
