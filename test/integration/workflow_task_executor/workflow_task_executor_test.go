package workflow_task_executor

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tuannh982/simple-workflows-go/internal/fn"
	"github.com/tuannh982/simple-workflows-go/pkg/dto"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/ptr"
	"testing"
	"time"
)

func TestWorkflowTaskExecutorCallActivity1(t *testing.T) {
	now := time.Now()
	input := &mockStruct{}
	inputBytes, err := dataConverter.Marshal(input)
	assert.NoError(t, err)
	executor := initExecutor(t)
	newEvents := []*history.HistoryEvent{
		{
			WorkflowExecutionStarted: &history.WorkflowExecutionStarted{
				Name:                     fn.GetFunctionName(mockWorkflow1),
				Version:                  "100",
				Input:                    inputBytes,
				WorkflowID:               "mock workflow ID",
				ParentWorkflowInfo:       nil,
				ScheduleToStartTimestamp: now.UnixMilli(),
			},
		},
	}
	mockTask := &task.WorkflowTask{
		WorkflowID:     "mock workflow ID",
		FetchTimestamp: now.UnixMilli(),
		OldEvents:      make([]*history.HistoryEvent, 0),
		NewEvents:      newEvents,
	}
	taskResult, err := executor.Execute(context.TODO(), mockTask)
	assert.NoError(t, err)
	activityCall := taskResult.PendingActivities[0]
	assert.Equal(t, int32(1), activityCall.TaskScheduledID)
	assert.Equal(t, fn.GetFunctionName(mockActivity1), activityCall.Name)
}

func TestWorkflowTaskExecutorNonDeterministicError1(t *testing.T) {
	now := time.Now()
	input := &mockStruct{}
	inputBytes, err := dataConverter.Marshal(input)
	assert.NoError(t, err)
	executor := initExecutor(t)
	newEvents := []*history.HistoryEvent{
		{
			WorkflowExecutionStarted: &history.WorkflowExecutionStarted{
				Name:                     fn.GetFunctionName(mockWorkflow1),
				Version:                  "100",
				Input:                    inputBytes,
				WorkflowID:               "mock workflow ID",
				ParentWorkflowInfo:       nil,
				ScheduleToStartTimestamp: now.UnixMilli(),
			},
		},
		{
			ActivityScheduled: &history.ActivityScheduled{
				TaskScheduledID: 1,
				Name:            fn.GetFunctionName(mockActivity2), // wrong activity name
				Input:           inputBytes,
			},
		},
	}
	mockTask := &task.WorkflowTask{
		WorkflowID:     "mock workflow ID",
		FetchTimestamp: now.UnixMilli(),
		OldEvents:      make([]*history.HistoryEvent, 0),
		NewEvents:      newEvents,
	}
	_, err = executor.Execute(context.TODO(), mockTask)
	assert.Error(t, err)
}

func TestWorkflowTaskExecutorNonDeterministicError2(t *testing.T) {
	now := time.Now()
	input := &mockStruct{}
	inputBytes, err := dataConverter.Marshal(input)
	assert.NoError(t, err)
	executor := initExecutor(t)
	newEvents := []*history.HistoryEvent{
		{
			WorkflowExecutionStarted: &history.WorkflowExecutionStarted{
				Name:                     fn.GetFunctionName(mockWorkflow1),
				Version:                  "100",
				Input:                    inputBytes,
				WorkflowID:               "mock workflow ID",
				ParentWorkflowInfo:       nil,
				ScheduleToStartTimestamp: now.UnixMilli(),
			},
		},
		{
			ActivityScheduled: &history.ActivityScheduled{
				TaskScheduledID: 1,
				Name:            fn.GetFunctionName(mockActivity2),
				Input:           inputBytes,
			},
		},
		{
			ActivityCompleted: &history.ActivityCompleted{
				TaskScheduledID: 2, // wrong task scheduled ID
				ExecutionResult: dto.ExecutionResult{
					Result: ptr.Ptr(inputBytes),
					Error:  nil,
				},
			},
		},
	}
	mockTask := &task.WorkflowTask{
		WorkflowID:     "mock workflow ID",
		FetchTimestamp: now.UnixMilli(),
		OldEvents:      make([]*history.HistoryEvent, 0),
		NewEvents:      newEvents,
	}
	_, err = executor.Execute(context.TODO(), mockTask)
	assert.Error(t, err)
}

func TestWorkflowTaskExecutorCallTimer(t *testing.T) {
	now := time.Now()
	input := &mockStruct{}
	inputBytes, err := dataConverter.Marshal(input)
	assert.NoError(t, err)
	executor := initExecutor(t)
	newEvents := []*history.HistoryEvent{
		{
			WorkflowExecutionStarted: &history.WorkflowExecutionStarted{
				Name:                     fn.GetFunctionName(mockWorkflow1),
				Version:                  "100",
				Input:                    inputBytes,
				WorkflowID:               "mock workflow ID",
				ParentWorkflowInfo:       nil,
				ScheduleToStartTimestamp: now.UnixMilli(),
			},
		},
		{
			Timestamp:           now.UnixMilli(),
			WorkflowTaskStarted: &history.WorkflowTaskStarted{},
		},
		// call mockActivity1
		{
			ActivityScheduled: &history.ActivityScheduled{
				TaskScheduledID: 1,
				Name:            fn.GetFunctionName(mockActivity1),
				Input:           inputBytes,
			},
		},
		{
			ActivityCompleted: &history.ActivityCompleted{
				TaskScheduledID: 1,
				ExecutionResult: dto.ExecutionResult{
					Result: ptr.Ptr(inputBytes),
					Error:  nil,
				},
			},
		},
	}
	mockTask := &task.WorkflowTask{
		WorkflowID:     "mock workflow ID",
		FetchTimestamp: now.UnixMilli(),
		OldEvents:      make([]*history.HistoryEvent, 0),
		NewEvents:      newEvents,
	}
	taskResult, err := executor.Execute(context.TODO(), mockTask)
	assert.NoError(t, err)
	timerCall := taskResult.PendingTimers[0]
	assert.Equal(t, int32(2), timerCall.TimerID)
	assert.Equal(t, now.UnixMilli()+50*time.Second.Milliseconds(), timerCall.FireAt)
}

func TestWorkflowTaskExecutorNonDeterministicError3(t *testing.T) {
	now := time.Now()
	input := &mockStruct{}
	inputBytes, err := dataConverter.Marshal(input)
	assert.NoError(t, err)
	executor := initExecutor(t)
	newEvents := []*history.HistoryEvent{
		{
			WorkflowExecutionStarted: &history.WorkflowExecutionStarted{
				Name:                     fn.GetFunctionName(mockWorkflow1),
				Version:                  "100",
				Input:                    inputBytes,
				WorkflowID:               "mock workflow ID",
				ParentWorkflowInfo:       nil,
				ScheduleToStartTimestamp: now.UnixMilli(),
			},
		},
		// call mockActivity1
		{
			ActivityScheduled: &history.ActivityScheduled{
				TaskScheduledID: 1,
				Name:            fn.GetFunctionName(mockActivity1),
				Input:           inputBytes,
			},
		},
		{
			ActivityCompleted: &history.ActivityCompleted{
				TaskScheduledID: 1,
				ExecutionResult: dto.ExecutionResult{
					Result: ptr.Ptr(inputBytes),
					Error:  nil,
				},
			},
		},
		// call timer
		{
			TimerCreated: &history.TimerCreated{
				TimerID: 2,
				FireAt:  now.UnixMilli() + 51*time.Second.Milliseconds(), // wrong fire at value
			},
		},
		{
			TimerFired: &history.TimerFired{TimerID: 2},
		},
	}
	mockTask := &task.WorkflowTask{
		WorkflowID:     "mock workflow ID",
		FetchTimestamp: now.UnixMilli(),
		OldEvents:      make([]*history.HistoryEvent, 0),
		NewEvents:      newEvents,
	}
	_, err = executor.Execute(context.TODO(), mockTask)
	assert.Error(t, err)
}

func TestWorkflowTaskExecutorCallActivity2(t *testing.T) {
	now := time.Now()
	input := &mockStruct{}
	inputBytes, err := dataConverter.Marshal(input)
	assert.NoError(t, err)
	executor := initExecutor(t)
	newEvents := []*history.HistoryEvent{
		{
			WorkflowExecutionStarted: &history.WorkflowExecutionStarted{
				Name:                     fn.GetFunctionName(mockWorkflow1),
				Version:                  "100",
				Input:                    inputBytes,
				WorkflowID:               "mock workflow ID",
				ParentWorkflowInfo:       nil,
				ScheduleToStartTimestamp: now.UnixMilli(),
			},
		},
		{
			Timestamp:           now.UnixMilli(),
			WorkflowTaskStarted: &history.WorkflowTaskStarted{},
		},
		// call mockActivity1
		{
			ActivityScheduled: &history.ActivityScheduled{
				TaskScheduledID: 1,
				Name:            fn.GetFunctionName(mockActivity1),
				Input:           inputBytes,
			},
		},
		{
			ActivityCompleted: &history.ActivityCompleted{
				TaskScheduledID: 1,
				ExecutionResult: dto.ExecutionResult{
					Result: ptr.Ptr(inputBytes),
					Error:  nil,
				},
			},
		},
		// call timer
		{
			TimerCreated: &history.TimerCreated{
				TimerID: 2,
				FireAt:  now.UnixMilli() + 50*time.Second.Milliseconds(),
			},
		},
		{
			TimerFired: &history.TimerFired{TimerID: 2},
		},
	}
	mockTask := &task.WorkflowTask{
		WorkflowID:     "mock workflow ID",
		FetchTimestamp: now.UnixMilli(),
		OldEvents:      make([]*history.HistoryEvent, 0),
		NewEvents:      newEvents,
	}
	taskResult, err := executor.Execute(context.TODO(), mockTask)
	assert.NoError(t, err)
	activityCall := taskResult.PendingActivities[0]
	assert.Equal(t, int32(3), activityCall.TaskScheduledID)
	assert.Equal(t, fn.GetFunctionName(mockActivity2), activityCall.Name)
}

func TestWorkflowTaskExecutorComplete(t *testing.T) {
	now := time.Now()
	input := &mockStruct{}
	inputBytes, err := dataConverter.Marshal(input)
	assert.NoError(t, err)
	executor := initExecutor(t)
	newEvents := []*history.HistoryEvent{
		{
			WorkflowExecutionStarted: &history.WorkflowExecutionStarted{
				Name:                     fn.GetFunctionName(mockWorkflow1),
				Version:                  "100",
				Input:                    inputBytes,
				WorkflowID:               "mock workflow ID",
				ParentWorkflowInfo:       nil,
				ScheduleToStartTimestamp: now.UnixMilli(),
			},
		},
		{
			Timestamp:           now.UnixMilli(),
			WorkflowTaskStarted: &history.WorkflowTaskStarted{},
		},
		// call mockActivity1
		{
			ActivityScheduled: &history.ActivityScheduled{
				TaskScheduledID: 1,
				Name:            fn.GetFunctionName(mockActivity1),
				Input:           inputBytes,
			},
		},
		{
			ActivityCompleted: &history.ActivityCompleted{
				TaskScheduledID: 1,
				ExecutionResult: dto.ExecutionResult{
					Result: ptr.Ptr(inputBytes),
					Error:  nil,
				},
			},
		},
		// call timer
		{
			TimerCreated: &history.TimerCreated{
				TimerID: 2,
				FireAt:  now.UnixMilli() + 50*time.Second.Milliseconds(),
			},
		},
		{
			TimerFired: &history.TimerFired{TimerID: 2},
		},
		// call mockActivity2
		{
			ActivityScheduled: &history.ActivityScheduled{
				TaskScheduledID: 3,
				Name:            fn.GetFunctionName(mockActivity2),
				Input:           inputBytes,
			},
		},
		{
			ActivityCompleted: &history.ActivityCompleted{
				TaskScheduledID: 3,
				ExecutionResult: dto.ExecutionResult{
					Result: nil,
					Error:  &dto.Error{Message: "error from mockActivity2"}, // return error
				},
			},
		},
	}
	mockTask := &task.WorkflowTask{
		WorkflowID:     "mock workflow ID",
		FetchTimestamp: now.UnixMilli(),
		OldEvents:      make([]*history.HistoryEvent, 0),
		NewEvents:      newEvents,
	}
	taskResult, err := executor.Execute(context.TODO(), mockTask)
	assert.NoError(t, err)
	assert.NotNil(t, taskResult)
	assert.NotNil(t, taskResult.WorkflowExecutionCompleted)
	workflowExecutionCompleted := taskResult.WorkflowExecutionCompleted
	assert.Nil(t, workflowExecutionCompleted.Result)
	assert.NotNil(t, workflowExecutionCompleted.Error)
	assert.Equal(t, "error from mockActivity2", workflowExecutionCompleted.Error.Message)
}
