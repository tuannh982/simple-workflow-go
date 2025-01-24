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
	TaskScheduleEvent *history.ActivityScheduled
}

type ActivityTaskExecutionError struct {
	Error             error
	NextExecutionTime *time.Time
}

type ActivityTaskResult struct {
	Task            *ActivityTask
	ExecutionResult *dto.ExecutionResult
	ExecutionError  *ActivityTaskExecutionError
}
