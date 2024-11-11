package history

type HistoryEvent struct {
	Timestamp int64 // millisecond precision

	// workflow start and end events
	WorkflowExecutionStarted    *WorkflowExecutionStarted    `json:",omitempty"`
	WorkflowExecutionCompleted  *WorkflowExecutionCompleted  `json:",omitempty"`
	WorkflowExecutionTerminated *WorkflowExecutionTerminated `json:",omitempty"`

	// workflow events when workflow worker picked/finished a workflow task
	WorkflowTaskStarted *WorkflowTaskStarted `json:",omitempty"`

	// activity events
	ActivityScheduled *ActivityScheduled `json:",omitempty"`
	ActivityCompleted *ActivityCompleted `json:",omitempty"`

	// timers
	TimerCreated *TimerCreated `json:",omitempty"`
	TimerFired   *TimerFired   `json:",omitempty"`

	// external events
	ExternalEventReceived *ExternalEventReceived `json:",omitempty"`
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
	} else if e.ExternalEventReceived != nil {
		return "ExternalEventReceived"
	}
	return "unknown"
}
