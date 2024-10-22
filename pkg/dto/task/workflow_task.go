package task

import (
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
)

type WorkflowTask struct {
	SeqNo          int32
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
