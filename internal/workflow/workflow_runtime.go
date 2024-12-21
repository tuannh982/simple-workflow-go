package workflow

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/tuannh982/simple-workflow-go/internal/fn"
	"github.com/tuannh982/simple-workflow-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflow-go/pkg/dto"
	"github.com/tuannh982/simple-workflow-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflow-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflow-go/pkg/registry"
	"runtime/debug"
)

var ErrControlledPanic = errors.New("controlled panic")
var ErrNonDeterministicError = errors.New("non-deterministic error")

type WorkflowRuntime struct {
	// init
	WorkflowRegistry *registry.WorkflowRegistry
	DataConverter    dataconverter.DataConverter
	// workflow task
	Task *task.WorkflowTask
	// runtime state
	HistoryIndex            int
	IsReplaying             bool
	SequenceNo              int64
	CurrentTimestamp        int64
	ActivityScheduledEvents map[int64]*history.ActivityScheduled
	ActivityPromises        map[int64]*UntypedActivityPromise
	TimerCreatedEvents      map[int64]*history.TimerCreated
	TimerPromises           map[int64]*TimerPromise
	// start
	WorkflowExecutionStartedTimestamp int64
	WorkflowExecutionStartedEvent     *history.WorkflowExecutionStarted
	WorkflowExecutionContext          *WorkflowExecutionContext
	Version                           string
	// final event, can only be execution completed event, and it's internally emit
	WorkflowExecutionCompleted *history.WorkflowExecutionCompleted
}

func NewWorkflowRuntime(
	workflowRegistry *registry.WorkflowRegistry,
	dataConverter dataconverter.DataConverter,
	task *task.WorkflowTask,
) *WorkflowRuntime {
	return &WorkflowRuntime{
		WorkflowRegistry:        workflowRegistry,
		DataConverter:           dataConverter,
		Task:                    task,
		HistoryIndex:            0,
		IsReplaying:             true,
		SequenceNo:              1, // avoid 0
		ActivityScheduledEvents: make(map[int64]*history.ActivityScheduled),
		ActivityPromises:        make(map[int64]*UntypedActivityPromise),
		TimerCreatedEvents:      make(map[int64]*history.TimerCreated),
		TimerPromises:           make(map[int64]*TimerPromise),
	}
}

func (w *WorkflowRuntime) GetWorkflowTaskResult() *task.WorkflowTaskResult {
	pendingActivities := make([]*history.ActivityScheduled, 0, len(w.ActivityScheduledEvents))
	for _, event := range w.ActivityScheduledEvents {
		pendingActivities = append(pendingActivities, event)
	}
	pendingTimers := make([]*history.TimerCreated, 0, len(w.TimerCreatedEvents))
	for _, event := range w.TimerCreatedEvents {
		pendingTimers = append(pendingTimers, event)
	}
	return &task.WorkflowTaskResult{
		Task:                       w.Task,
		PendingActivities:          pendingActivities,
		PendingTimers:              pendingTimers,
		WorkflowExecutionCompleted: w.WorkflowExecutionCompleted,
	}
}

func (w *WorkflowRuntime) RunSimulation() (err error) {
	defer func() {
		if r := recover(); r != nil {
			if r != ErrControlledPanic {
				err = fmt.Errorf("panic: %v\nstack: %s", r, string(debug.Stack()))
			}
		}
	}()
	for {
		done, unExpectedErr := w.Step()
		if unExpectedErr != nil {
			return unExpectedErr
		}
		if done {
			break
		}
	}
	return nil
}

func (w *WorkflowRuntime) Step() (bool, error) {
	event, done := w.nextHistoryEvent()
	if done {
		return true, nil
	}
	err := w.processEvent(event)
	return false, err
}

func (w *WorkflowRuntime) nextHistoryEvent() (*history.HistoryEvent, bool) {
	var tape []*history.HistoryEvent
	var tapeIndex int
	if w.HistoryIndex < len(w.Task.OldEvents) {
		w.IsReplaying = true
		tape = w.Task.OldEvents
		tapeIndex = w.HistoryIndex
	} else if w.HistoryIndex < len(w.Task.OldEvents)+len(w.Task.NewEvents) {
		w.IsReplaying = false
		tape = w.Task.NewEvents
		tapeIndex = w.HistoryIndex - len(w.Task.OldEvents)
	} else {
		return nil, true
	}
	w.HistoryIndex++
	return tape[tapeIndex], false
}

func (w *WorkflowRuntime) processEvent(event *history.HistoryEvent) error {
	var err error
	if e := event.WorkflowExecutionStarted; e != nil {
		err = w.handleWorkflowExecutionStarted(event)
	} else if e := event.WorkflowExecutionCompleted; e != nil {
		err = w.handleWorkflowExecutionCompleted(event)
	} else if e := event.WorkflowExecutionTerminated; e != nil {
		err = w.handleWorkflowExecutionTerminated(event)
	} else if e := event.WorkflowTaskStarted; e != nil {
		err = w.handleWorkflowTaskStarted(event)
	} else if e := event.ActivityScheduled; e != nil {
		err = w.handleActivityScheduled(event)
	} else if e := event.ActivityCompleted; e != nil {
		err = w.handleActivityCompleted(event)
	} else if e := event.TimerCreated; e != nil {
		err = w.handleTimerCreated(event)
	} else if e := event.TimerFired; e != nil {
		err = w.handleTimerFired(event)
	} else if e := event.ExternalEventReceived; e != nil {
		err = w.handleExternalEventReceived(event)
	} else {
		err = fmt.Errorf("don't know how to handle event %v", event)
	}
	return err
}

/*
	Workflow execution
*/

func (w *WorkflowRuntime) executeWorkflow(
	workflow any,
	ctx context.Context,
	input any,
) (workflowExecutionCompleted *history.WorkflowExecutionCompleted, err error) {
	callResult, callErr := fn.CallFn(workflow, ctx, input)
	marshaledCallResult, err := dto.ExtractResultFromFnCallResult(callResult, w.DataConverter.Marshal)
	if err != nil {
		return nil, err
	}
	wrappedCallError := dto.ExtractErrorFromFnCallError(callErr)
	return &history.WorkflowExecutionCompleted{
		ExecutionResult: dto.ExecutionResult{
			Result: marshaledCallResult,
			Error:  wrappedCallError,
		},
	}, nil
}

func (w *WorkflowRuntime) handleWorkflowExecutionStarted(event *history.HistoryEvent) error {
	if e := event.WorkflowExecutionStarted; e != nil {
		if w.WorkflowExecutionStartedEvent != nil {
			// ignore duplicated event
			return nil
		}
		w.WorkflowExecutionStartedTimestamp = event.Timestamp
		w.CurrentTimestamp = event.Timestamp
		w.WorkflowExecutionStartedEvent = event.WorkflowExecutionStarted
		w.Version = event.WorkflowExecutionStarted.Version
		name := e.Name
		inputBytes := e.Input
		if workflow, ok := w.WorkflowRegistry.Workflows[name]; ok {
			w.WorkflowExecutionContext = NewWorkflowExecutionContext(w)
			ctx := InjectWorkflowExecutionContext(context.Background(), w.WorkflowExecutionContext)
			input := fn.InitArgument(workflow)
			err := w.DataConverter.Unmarshal(inputBytes, input)
			if err != nil {
				return err
			}
			workflowExecutionCompleted, err := w.executeWorkflow(workflow, ctx, input)
			if err != nil {
				return err
			}
			w.WorkflowExecutionCompleted = workflowExecutionCompleted
			return nil
		} else {
			return fmt.Errorf("workflow %s not found", name)
		}
	} else {
		return fmt.Errorf("expect WorkflowExecutionStarted event, got %v", e)
	}
}

func (w *WorkflowRuntime) handleWorkflowExecutionCompleted(event *history.HistoryEvent) error {
	if e := event.WorkflowExecutionCompleted; e != nil {
		// do nothing
		return nil
	} else {
		return fmt.Errorf("expect WorkflowExecutionCompleted event, got %v", e)
	}
}

func (w *WorkflowRuntime) handleWorkflowExecutionTerminated(event *history.HistoryEvent) error {
	if e := event.WorkflowExecutionTerminated; e != nil {
		// do nothing
		return nil
	} else {
		return fmt.Errorf("expect WorkflowExecutionTerminated event, got %v", e)
	}
}

/*
	Workflow task execution
*/

func (w *WorkflowRuntime) handleWorkflowTaskStarted(event *history.HistoryEvent) error {
	if e := event.WorkflowTaskStarted; e != nil {
		w.CurrentTimestamp = event.Timestamp
		return nil
	} else {
		return fmt.Errorf("expect WorkflowTaskStarted event, got %v", e)
	}
}

/*
	Activity
*/

func (w *WorkflowRuntime) ScheduleNewActivity(activity any, input any) *ActivityPromise {
	taskScheduledID := w.nextSeqNo()
	name := fn.GetFunctionName(activity)
	if w.ActivityPromises[taskScheduledID] != nil {
		return w.ActivityPromises[taskScheduledID].ToTyped(activity)
	}
	promise := NewUntypedActivityPromise(w)
	inputBytes, err := w.DataConverter.Marshal(input)
	if err != nil {
		panic(err)
	}
	event := &history.ActivityScheduled{
		TaskScheduledID: taskScheduledID,
		Name:            name,
		Input:           inputBytes,
	}
	// this part is implement differently with DTFx, this allows some out-of-order task schedule event
	if e, ok := w.ActivityScheduledEvents[taskScheduledID]; ok {
		if e.TaskScheduledID != event.TaskScheduledID || e.Name != event.Name || !bytes.Equal(e.Input, event.Input) {
			panic(ErrNonDeterministicError)
		}
	} else {
		w.ActivityScheduledEvents[taskScheduledID] = event
	}
	w.ActivityPromises[taskScheduledID] = promise
	return promise.ToTyped(activity)
}

func (w *WorkflowRuntime) handleActivityScheduled(event *history.HistoryEvent) error {
	if e := event.ActivityScheduled; e != nil {
		if activityScheduled, ok := w.ActivityScheduledEvents[e.TaskScheduledID]; ok {
			if activityScheduled.Name != e.Name || !bytes.Equal(activityScheduled.Input, e.Input) {
				return ErrNonDeterministicError
			}
			delete(w.ActivityScheduledEvents, e.TaskScheduledID)
		} else {
			// in some cases, when using defer with await, the events might not be correct in order, but the code is still correct,
			// so we will insert an event without promise here
			w.ActivityScheduledEvents[e.TaskScheduledID] = e
		}
		return nil
	} else {
		return fmt.Errorf("expect ActivityScheduled event, got %v", e)
	}
}

func (w *WorkflowRuntime) handleActivityCompleted(event *history.HistoryEvent) error {
	if e := event.ActivityCompleted; e != nil {
		promise, ok := w.ActivityPromises[e.TaskScheduledID]
		if !ok {
			promise = NewUntypedActivityPromise(w)
			w.ActivityPromises[e.TaskScheduledID] = promise
		}
		if e.Error != nil {
			promise.Promise.Reject(errors.New(e.Error.Message))
		} else {
			promise.Promise.Resolve(e.Result)
		}
		return nil
	} else {
		return fmt.Errorf("expect ActivityCompleted event, got %v", e)
	}
}

/*
	Timer
*/

func (w *WorkflowRuntime) CreateTimer(fireAt int64) *TimerPromise {
	promise := NewTimerPromise(w)
	timerID := w.nextSeqNo()
	event := &history.TimerCreated{
		TimerID: timerID,
		FireAt:  fireAt,
	}
	// this part is implement differently with DTFx, this allows some out-of-order task schedule event
	if e, ok := w.TimerCreatedEvents[timerID]; ok {
		if e.TimerID != event.TimerID || e.FireAt != event.FireAt {
			panic(ErrNonDeterministicError)
		}
	} else {
		w.TimerCreatedEvents[timerID] = event
	}
	w.TimerPromises[timerID] = promise
	return promise
}

func (w *WorkflowRuntime) handleTimerCreated(event *history.HistoryEvent) error {
	if e := event.TimerCreated; e != nil {
		if timerCreated, ok := w.TimerCreatedEvents[e.TimerID]; ok {
			if timerCreated.FireAt != e.FireAt {
				return ErrNonDeterministicError
			}
			delete(w.TimerCreatedEvents, e.TimerID)
		} else {
			// in some cases, when using defer with await, the events might not be correct in order, but the code is still correct,
			// so we will insert an event without promise here
			w.TimerCreatedEvents[e.TimerID] = e
		}
		return nil
	} else {
		return fmt.Errorf("expect TimerCreated event, got %v", e)
	}
}

func (w *WorkflowRuntime) handleTimerFired(event *history.HistoryEvent) error {
	if e := event.TimerFired; e != nil {
		promise, ok := w.TimerPromises[e.TimerID]
		if !ok {
			promise = NewTimerPromise(w)
			w.TimerPromises[e.TimerID] = promise
		}
		promise.Promise.Resolve(&struct{}{})
		return nil
	} else {
		return fmt.Errorf("expect TimerFired event, got %v", e)
	}
}

func (w *WorkflowRuntime) handleExternalEventReceived(event *history.HistoryEvent) error {
	if e := event.ExternalEventReceived; e != nil {
		callbackRegistry := w.WorkflowExecutionContext.EventCallbacks
		if callbacks, ok := callbackRegistry[e.EventName]; ok {
			for _, callback := range callbacks {
				callback(e.Input)
			}
		}
		return nil
	} else {
		return fmt.Errorf("expect ExternalEventReceived event, got %v", e)
	}
}

/*
	Sequence number
*/

func (w *WorkflowRuntime) nextSeqNo() int64 {
	result := w.SequenceNo
	w.SequenceNo++
	return result
}
