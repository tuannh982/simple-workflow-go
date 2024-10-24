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
