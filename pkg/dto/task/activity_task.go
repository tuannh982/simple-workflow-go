package task

import (
	"github.com/tuannh982/simple-workflows-go/pkg/dto"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
)

type ActivityTask struct {
	WorkflowID        string
	TaskScheduleEvent *history.ActivityScheduled
}

type ActivityTaskResult struct {
	Task            *ActivityTask
	ExecutionResult *dto.ExecutionResult
}
