package task

import (
	"github.com/tuannh982/simple-workflows-go/internal/dto"
	"github.com/tuannh982/simple-workflows-go/internal/dto/history"
)

type ActivityTask struct {
	WorkflowID        string
	TaskScheduleEvent *history.ActivityScheduled
}

type ActivityTaskResult struct {
	Task            *ActivityTask
	ExecutionResult *dto.ExecutionResult
}
