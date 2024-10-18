package task

import "github.com/tuannh982/simple-workflows-go/internal/dto/history"

type WorkflowTask struct {
	WorkflowID     string
	FetchTimestamp int64
	OldEvents      []*history.HistoryEvent
	NewEvents      []*history.HistoryEvent
}

type WorkflowTaskResult struct {
	Task                       *WorkflowTask
	PendingActivities          []*history.ActivityScheduled
	PendingTimers              []*history.TimerCreated
	WorkflowExecutionCompleted *history.WorkflowExecutionCompleted
}
