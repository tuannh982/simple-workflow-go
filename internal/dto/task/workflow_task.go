package task

import "github.com/tuannh982/simple-workflows-go/internal/dto/history"

type WorkflowTask struct {
	WorkflowID string
	OldEvents  []*history.HistoryEvent
	NewEvents  []*history.Event
}

type WorkflowTaskResult struct {
	Task              *WorkflowTask
	NewEvents         []*history.Event
	PendingActivities []*history.ActivityScheduled
	PendingTimers     []*history.TimerCreated
}
