package workflow

import (
	"context"
	"errors"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/internal/dataconverter"
	"github.com/tuannh982/simple-workflows-go/internal/dto"
	"github.com/tuannh982/simple-workflows-go/internal/dto/history"
	"github.com/tuannh982/simple-workflows-go/internal/dto/task"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
)

var ErrControlledPanic = errors.New("controlled panic")

type WorkflowRuntime struct {
	// init
	WorkflowRegistry *WorkflowRegistry
	DataConverter    dataconverter.DataConverter
	// workflow task
	Task *task.WorkflowTask
	// runtime state
	HistoryIndex            int
	IsReplaying             bool
	SequenceNo              int32
	CurrentTimestamp        int64
	ActivityScheduledEvents map[int32]*history.ActivityScheduled
	ActivityPromises        map[int32]*ActivityPromise
	TimerCreatedEvents      map[int32]*history.TimerCreated
	TimerPromises           map[int32]*TimerPromise
	// start event
	WorkflowExecutionStartedEvent *history.HistoryEvent
	// final event, can only be execution completed event, and it's internally emit
	WorkflowExecutionCompleted *history.WorkflowExecutionCompleted
}

func NewWorkflowRuntime(
	workflowRegistry *WorkflowRegistry,
	dataConverter dataconverter.DataConverter,
	task *task.WorkflowTask,
) *WorkflowRuntime {
	return &WorkflowRuntime{
		WorkflowRegistry:        workflowRegistry,
		DataConverter:           dataConverter,
		Task:                    task,
		HistoryIndex:            0,
		IsReplaying:             true,
		SequenceNo:              0,
		ActivityScheduledEvents: make(map[int32]*history.ActivityScheduled),
		ActivityPromises:        make(map[int32]*ActivityPromise),
		TimerCreatedEvents:      make(map[int32]*history.TimerCreated),
		TimerPromises:           make(map[int32]*TimerPromise),
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
				err = fmt.Errorf("panic: %v", r)
			}
		}
	}()
	for {
		done, unexpectedErr := w.Step()
		if unexpectedErr != nil {
			return unexpectedErr
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
	} else {
		err = fmt.Errorf("don't know how to handle event %v", event)
	}
	return err
}

/*
	Workflow execution
*/

func (w *WorkflowRuntime) executeWorkflow(workflow any, ctx context.Context, input any) (*history.WorkflowExecutionCompleted, error) {
	var err error
	workflowResult, workflowErr := fn.CallFn(workflow, ctx, input)
	var marshaledWorkflowResult *[]byte
	var wrappedWorkflowError *dto.Error
	if workflowErr != nil {
		wrappedWorkflowError = &dto.Error{
			Message: workflowErr.Error(),
		}
	}
	if workflowResult != nil {
		var r []byte
		r, err = w.DataConverter.Marshal(workflowResult)
		marshaledWorkflowResult = &r
		if err != nil {
			return nil, err
		}
	}
	return &history.WorkflowExecutionCompleted{
		ExecutionResult: dto.ExecutionResult{
			Result: marshaledWorkflowResult,
			Error:  wrappedWorkflowError,
		},
	}, nil
}

func (w *WorkflowRuntime) handleWorkflowExecutionStarted(event *history.HistoryEvent) error {
	if e := event.WorkflowExecutionStarted; e != nil {
		w.WorkflowExecutionStartedEvent = event
		name := e.Name
		inputBytes := e.Input
		if workflow, ok := w.WorkflowRegistry.workflows[name]; ok {
			ctx := InjectWorkflowExecutionContext(context.Background(), NewWorkflowExecutionContext(w))
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
	promise := NewActivityPromise(w)
	taskScheduledID := w.nextSeqNo()
	name := fn.GetFunctionName(activity)
	bytes, err := w.DataConverter.Marshal(input)
	if err != nil {
		panic(err)
	}
	event := &history.ActivityScheduled{
		TaskScheduledID: taskScheduledID,
		Name:            name,
		Input:           bytes,
	}
	w.ActivityScheduledEvents[taskScheduledID] = event
	w.ActivityPromises[taskScheduledID] = promise
	return promise
}

func (w *WorkflowRuntime) handleActivityScheduled(event *history.HistoryEvent) error {
	if e := event.ActivityScheduled; e != nil {
		if _, ok := w.ActivityScheduledEvents[e.TaskScheduledID]; !ok {
			return fmt.Errorf("activity scheduled event not found %v", e)
		}
		delete(w.ActivityScheduledEvents, e.TaskScheduledID)
		return nil
	} else {
		return fmt.Errorf("expect ActivityScheduled event, got %v", e)
	}
}

func (w *WorkflowRuntime) handleActivityCompleted(event *history.HistoryEvent) error {
	if e := event.ActivityCompleted; e != nil {
		promise, ok := w.ActivityPromises[e.TaskScheduledID]
		if !ok {
			// TODO handle duplicated event?
			return nil
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
	w.TimerCreatedEvents[timerID] = event
	w.TimerPromises[timerID] = promise
	return promise
}

func (w *WorkflowRuntime) handleTimerCreated(event *history.HistoryEvent) error {
	if e := event.TimerCreated; e != nil {
		if _, ok := w.TimerCreatedEvents[e.TimerID]; !ok {
			return fmt.Errorf("timer created event not found %v", e)
		}
		delete(w.TimerCreatedEvents, e.TimerID)
		return nil
	} else {
		return fmt.Errorf("expect TimerCreated event, got %v", e)
	}
}

func (w *WorkflowRuntime) handleTimerFired(event *history.HistoryEvent) error {
	if e := event.TimerFired; e != nil {
		promise, ok := w.TimerPromises[e.TimerID]
		if !ok {
			// TODO handle duplicated event?
			return nil
		}
		promise.Promise.Resolve(&struct{}{})
		return nil
	} else {
		return fmt.Errorf("expect TimerFired event, got %v", e)
	}
}

/*
	Sequence number
*/

func (w *WorkflowRuntime) nextSeqNo() int32 {
	result := w.SequenceNo
	w.SequenceNo++
	return result
}
