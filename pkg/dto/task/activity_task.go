package task

import (
	"github.com/tuannh982/simple-workflow-go/pkg/dto"
	"github.com/tuannh982/simple-workflow-go/pkg/dto/history"
)

type ActivityTask struct {
	TaskID            string
	WorkflowID        string
	NumAttempted      int
	TaskScheduleEvent *history.ActivityScheduled
}

type ActivityTaskResult struct {
	Task            *ActivityTask
	ExecutionResult *dto.ExecutionResult
}
