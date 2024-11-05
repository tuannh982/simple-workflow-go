package history

type HistoryEvent struct {
	Timestamp int64 // millisecond precision

	// workflow start and end events
	WorkflowExecutionStarted    *WorkflowExecutionStarted
	WorkflowExecutionCompleted  *WorkflowExecutionCompleted
	WorkflowExecutionTerminated *WorkflowExecutionTerminated

	// workflow events when workflow worker picked/finished a workflow task
	WorkflowTaskStarted *WorkflowTaskStarted

	// activity events
	ActivityScheduled *ActivityScheduled
	ActivityCompleted *ActivityCompleted

	// timers
	TimerCreated *TimerCreated
	TimerFired   *TimerFired
}

func (e *HistoryEvent) GetType() string {
	if e.WorkflowExecutionStarted != nil {
		return "WorkflowExecutionStarted"
	} else if e.WorkflowExecutionCompleted != nil {
		return "WorkflowExecutionCompleted"
	} else if e.WorkflowExecutionTerminated != nil {
		return "WorkflowExecutionTerminated"
	} else if e.WorkflowTaskStarted != nil {
		return "WorkflowTaskStarted"
	} else if e.ActivityScheduled != nil {
		return "ActivityScheduled"
	} else if e.ActivityCompleted != nil {
		return "ActivityCompleted"
	} else if e.TimerCreated != nil {
		return "TimerCreated"
	} else if e.TimerFired != nil {
		return "TimerFired"
	}
	return "unknown"
}
