package task

import (
	"github.com/tuannh982/simple-workflow-go/pkg/dto/history"
)

type WorkflowTask struct {
	TaskID         string
	WorkflowID     string
	FetchTimestamp int64
	OldEvents      []*history.HistoryEvent
	NewEvents      []*history.HistoryEvent
}

type WorkflowTaskResult struct {
	Task                       *WorkflowTask
	PendingActivities          []*history.ActivityScheduled
	PendingTimers              []*history.TimerCreated
	WorkflowExecutionCompleted *history.WorkflowExecutionCompleted `json:",omitempty"`
}
