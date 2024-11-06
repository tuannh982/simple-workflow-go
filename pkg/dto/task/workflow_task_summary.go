package task

import (
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/collections"
	"time"
)

type WorkflowTaskSummary struct {
	TaskID         string
	WorkflowID     string
	FetchTimestamp time.Time
	OldEvents      []string
	NewEvents      []string
}

func (t *WorkflowTask) Summary() any {
	oldEvents := collections.MapArray(t.OldEvents, func(e *history.HistoryEvent) string { return e.GetType() })
	newEvents := collections.MapArray(t.NewEvents, func(e *history.HistoryEvent) string { return e.GetType() })
	return &WorkflowTaskSummary{
		TaskID:         t.TaskID,
		WorkflowID:     t.WorkflowID,
		FetchTimestamp: time.UnixMilli(t.FetchTimestamp),
		OldEvents:      oldEvents,
		NewEvents:      newEvents,
	}
}

type WorkflowTaskResultSummary struct {
	TaskID                     string
	WorkflowID                 string
	TaskFetchTimestamp         time.Time
	PendingActivities          []*history.ActivityScheduled
	PendingTimers              []*history.TimerCreated
	WorkflowExecutionCompleted *history.WorkflowExecutionCompleted `json:",omitempty"`
}

func (r *WorkflowTaskResult) Summary() any {
	return &WorkflowTaskResultSummary{
		TaskID:                     r.Task.TaskID,
		WorkflowID:                 r.Task.WorkflowID,
		TaskFetchTimestamp:         time.UnixMilli(r.Task.FetchTimestamp),
		PendingActivities:          r.PendingActivities,
		PendingTimers:              r.PendingTimers,
		WorkflowExecutionCompleted: r.WorkflowExecutionCompleted,
	}
}
