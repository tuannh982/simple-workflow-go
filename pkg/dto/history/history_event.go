package history

// TODO should separate into 2 structs: Event and HistoryEvent since Event is only has Timestamp and EventID after becoming HistoryEvent
type HistoryEvent struct {
	EventID   int32
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
