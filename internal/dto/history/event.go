package history

type Event struct {
	// workflow start and end events
	WorkflowExecutionStarted    *WorkflowExecutionStarted
	WorkflowExecutionCompleted  *WorkflowExecutionCompleted
	WorkflowExecutionTerminated *WorkflowExecutionTerminated

	// workflow events when workflow worker picked/finished a workflow task
	WorkflowTaskStarted   *WorkflowTaskStarted
	WorkflowTaskCompleted *WorkflowTaskCompleted

	// activity events
	ActivityScheduled *ActivityScheduled
	ActivityCompleted *ActivityCompleted

	// timers
	TimerCreated *TimerCreated
	TimerFired   *TimerFired
}
