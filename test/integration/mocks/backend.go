package mocks

import (
	"context"
	"errors"
	"fmt"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/dto"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/collections"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/ptr"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/worker"
	"os"
	"sort"
	"sync"
	"time"
)

type mockBackend struct {
	dataConverter  dataconverter.DataConverter
	persistent     *MockPersistent
	thisInstanceID string
	*sync.Mutex
}

func NewMockBackend(dataConverter dataconverter.DataConverter) backend.Backend {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	now := time.Now().UnixMilli()
	thisInstanceID := fmt.Sprintf("%s_%d", hostname, now)
	return &mockBackend{
		dataConverter:  dataConverter,
		persistent:     NewMockPersistent(),
		thisInstanceID: thisInstanceID,
		Mutex:          &sync.Mutex{},
	}
}

func (m *mockBackend) DataConverter() dataconverter.DataConverter {
	return m.dataConverter
}

func (m *mockBackend) CreateWorkflow(_ context.Context, info *history.WorkflowExecutionStarted) error {
	m.Lock()
	defer m.Unlock()
	now := time.Now().UnixMilli()
	var parentWorkflowID *string
	if info.ParentWorkflowInfo != nil {
		parentWorkflowID = &info.ParentWorkflowInfo.WorkflowID
	}
	workflow := &MockDbWorkflow{
		id:                   info.WorkflowID,
		name:                 info.Name,
		version:              info.Version,
		createdAt:            now,
		currentRuntimeStatus: string(dto.WorkflowRuntimeStatusPending),
		input:                info.Input,
		parentWorkflowID:     parentWorkflowID,
	}
	workflowTask := &MockDbTask{
		sequenceNo: m.persistent.NextSeqNo(),
		workflowID: info.WorkflowID,
		taskType:   string(task.TaskTypeWorkflow),
		createdAt:  now,
		visibleAt:  info.ScheduleToStartTimestamp,
		payload:    info.Input,
	}
	payload, err := dto.Marshal(&history.HistoryEvent{
		Timestamp:                now,
		WorkflowExecutionStarted: info,
	})
	if err != nil {
		return err
	}
	event := &MockDbEvent{
		sequenceNo: m.persistent.NextSeqNo(),
		workflowID: info.WorkflowID,
		createdAt:  now,
		visibleAt:  info.ScheduleToStartTimestamp,
		payload:    payload,
	}
	m.persistent.InsertWorkflow(workflow)
	m.persistent.InsertTask(workflowTask)
	m.persistent.InsertEvent(event)
	return nil
}

func (m *mockBackend) GetWorkflowResult(_ context.Context, name string, workflowID string) (*dto.WorkflowExecutionResult, error) {
	m.Lock()
	defer m.Unlock()
	w := m.persistent.GetWorkflow(workflowID)
	if w == nil {
		return nil, errors.New("workflow not found")
	}
	if w.name != name {
		return nil, errors.New("wrong workflow type")
	}
	executionResult := dto.ExecutionResult{
		Result: w.resultOutput,
		Error:  collections.Map(w.resultError, func(a string) dto.Error { return dto.Error{Message: a} }),
	}
	return &dto.WorkflowExecutionResult{
		WorkflowID:      w.id,
		Version:         w.version,
		RuntimeStatus:   w.currentRuntimeStatus,
		ExecutionResult: executionResult,
	}, nil
}

func (m *mockBackend) AppendWorkflowEvent(_ context.Context, workflowID string, event *history.HistoryEvent) error {
	m.Lock()
	defer m.Unlock()
	now := time.Now().UnixMilli()
	payload, err := dto.Marshal(event)
	if err != nil {
		return err
	}
	e := &MockDbEvent{
		sequenceNo: m.persistent.NextSeqNo(),
		workflowID: workflowID,
		createdAt:  now,
		visibleAt:  now,
		payload:    payload,
	}
	m.persistent.InsertEvent(e)
	return nil
}

func (m *mockBackend) GetWorkflowTask(_ context.Context) (*task.WorkflowTask, error) {
	m.Lock()
	defer m.Unlock()
	now := time.Now().UnixMilli()
	// get workflow that has pending events
	availableTasks := m.persistent.FilterTasks(func(t *MockDbTask) bool {
		return t.taskType == string(task.TaskTypeWorkflow) &&
			t.visibleAt < now &&
			(t.lockedBy == nil || *t.lockedBy == m.thisInstanceID)

	})
	availableTasksMap := collections.ToMap(func(t *MockDbTask) string { return t.workflowID }, availableTasks)
	availableWorkflowEvents := m.persistent.FilterWorkflowEvents(func(e *MockDbEvent) bool {
		if e.visibleAt < now {
			_, ok := availableTasksMap[e.workflowID]
			return ok
		} else {
			return false
		}
	})
	availableWorkflowEventsGroupedByWorkflowID := collections.GroupBy(func(t *MockDbEvent) string { return t.workflowID }, availableWorkflowEvents)
	if len(availableWorkflowEventsGroupedByWorkflowID) == 0 {
		return nil, worker.ErrNoTask
	}
	selected, events := collections.FirstInMap(availableWorkflowEventsGroupedByWorkflowID)
	for _, event := range events {
		event.lockedBy = &m.thisInstanceID
	}
	workflowTask := availableTasksMap[selected]
	sort.Slice(events, func(i, j int) bool {
		return events[i].sequenceNo < events[j].sequenceNo
	})
	workflowHistoryEvents := m.persistent.GetWorkflowHistory(selected)
	sort.Slice(workflowHistoryEvents, func(i, j int) bool {
		return workflowHistoryEvents[i].sequenceNo < workflowHistoryEvents[j].sequenceNo
	})
	// craft task
	oldEvents := collections.MapArray(workflowHistoryEvents, func(a *MockDbHistoryEvent) *history.HistoryEvent {
		return m.ForceUnmarshalHistoryEvent(a.payload)
	})
	newEvents := collections.MapArray(events, func(a *MockDbEvent) *history.HistoryEvent {
		return m.ForceUnmarshalHistoryEvent(a.payload)
	})
	// lock task
	workflowTask.lockedBy = &m.thisInstanceID
	// return
	return &task.WorkflowTask{
		TaskID:         fmt.Sprintf("%d", workflowTask.sequenceNo),
		WorkflowID:     selected,
		FetchTimestamp: now,
		OldEvents:      oldEvents,
		NewEvents:      newEvents,
	}, nil
}

func (m *mockBackend) CompleteWorkflowTask(_ context.Context, result *task.WorkflowTaskResult) error {
	m.Lock()
	defer m.Unlock()
	now := time.Now().UnixMilli()
	if t := m.persistent.GetTask(result.Task.TaskID); t != nil {
		if t.taskType == string(task.TaskTypeWorkflow) && *t.lockedBy == m.thisInstanceID {
			isCompleted := false
			processedEvents := result.Task.NewEvents
			// update workflow state
			workflow := m.persistent.GetWorkflow(result.Task.WorkflowID)
			for _, event := range processedEvents {
				if event.WorkflowExecutionStarted != nil {
					workflow.startAt = &event.Timestamp
					workflow.currentRuntimeStatus = string(dto.WorkflowRuntimeStatusRunning)
				} else if event.WorkflowExecutionCompleted != nil {
					workflow.completedAt = &event.Timestamp
					workflow.currentRuntimeStatus = string(dto.WorkflowRuntimeStatusCompleted)
					workflow.resultOutput = event.WorkflowExecutionCompleted.Result
					if event.WorkflowExecutionCompleted.Error != nil {
						workflow.resultError = ptr.Ptr(event.WorkflowExecutionCompleted.Error.Message)
					}
					isCompleted = true
				}
			}
			// delete processed events, and move them to history
			m.persistent.DeleteEventsByWorkflowAndLock(result.Task.WorkflowID, m.thisInstanceID)
			for _, event := range result.Task.NewEvents {
				m.persistent.InsertHistoryEvent(&MockDbHistoryEvent{
					workflowID: result.Task.WorkflowID,
					sequenceNo: m.persistent.NextSeqNo(),
					payload:    m.ForceMarshalHistoryEvent(event),
				})
			}
			// build new events list
			for _, tsk := range result.PendingActivities {
				b, err := dto.Marshal(tsk)
				if err != nil {
					return err
				}
				m.persistent.InsertTask(&MockDbTask{
					sequenceNo: m.persistent.NextSeqNo(),
					workflowID: result.Task.WorkflowID,
					taskType:   string(task.TaskTypeActivity),
					createdAt:  now,
					visibleAt:  now,
					payload:    b,
				})
				m.persistent.InsertEvent(&MockDbEvent{
					sequenceNo: m.persistent.NextSeqNo(),
					workflowID: result.Task.WorkflowID,
					createdAt:  now,
					visibleAt:  now,
					payload: m.ForceMarshalHistoryEvent(&history.HistoryEvent{
						Timestamp:         now,
						ActivityScheduled: tsk,
					}),
				})
			}
			for _, event := range result.PendingTimers {
				m.persistent.InsertEvent(&MockDbEvent{
					sequenceNo: m.persistent.NextSeqNo(),
					workflowID: result.Task.WorkflowID,
					createdAt:  now,
					visibleAt:  now,
					payload: m.ForceMarshalHistoryEvent(&history.HistoryEvent{
						Timestamp:    now,
						TimerCreated: event,
					}),
				})
				m.persistent.InsertEvent(&MockDbEvent{
					sequenceNo: m.persistent.NextSeqNo(),
					workflowID: result.Task.WorkflowID,
					createdAt:  now,
					visibleAt:  event.FireAt,
					payload: m.ForceMarshalHistoryEvent(&history.HistoryEvent{
						Timestamp:  event.FireAt,
						TimerFired: &history.TimerFired{TimerID: event.TimerID},
					}),
				})
			}
			if result.WorkflowExecutionCompleted != nil {
				m.persistent.InsertEvent(&MockDbEvent{
					sequenceNo: m.persistent.NextSeqNo(),
					workflowID: result.Task.WorkflowID,
					createdAt:  now,
					visibleAt:  now,
					payload: m.ForceMarshalHistoryEvent(&history.HistoryEvent{
						Timestamp:                  now,
						WorkflowExecutionCompleted: result.WorkflowExecutionCompleted,
					}),
				})
			}
			if isCompleted {
				m.persistent.RemoveTask(result.Task.TaskID)
			}
			return nil
		}
	}
	return errors.New("unexpected error")
}

func (m *mockBackend) AbandonWorkflowTask(_ context.Context, t *task.WorkflowTask, _ *string) error {
	m.Lock()
	defer m.Unlock()
	// unlock task
	if t := m.persistent.GetTask(t.TaskID); t != nil {
		if t.taskType == string(task.TaskTypeWorkflow) && *t.lockedBy == m.thisInstanceID {
			t.lockedBy = nil
		}
	}
	return nil
}

func (m *mockBackend) GetActivityTask(_ context.Context) (*task.ActivityTask, error) {
	m.Lock()
	defer m.Unlock()
	now := time.Now().UnixMilli()
	availableTasks := m.persistent.FilterTasks(func(t *MockDbTask) bool {
		return t.taskType == string(task.TaskTypeActivity) &&
			t.visibleAt < now &&
			(t.lockedBy == nil || *t.lockedBy == m.thisInstanceID)

	})
	if len(availableTasks) == 0 {
		return nil, worker.ErrNoTask
	}
	selected := collections.FirstInArray(availableTasks)
	selected.lockedBy = &m.thisInstanceID
	event := &history.ActivityScheduled{}
	err := dto.Unmarshal(selected.payload, event)
	if err != nil {
		return nil, err
	}
	return &task.ActivityTask{
		TaskID:            fmt.Sprintf("%d", selected.sequenceNo),
		WorkflowID:        selected.workflowID,
		TaskScheduleEvent: event,
	}, nil
}

func (m *mockBackend) CompleteActivityTask(_ context.Context, result *task.ActivityTaskResult) error {
	m.Lock()
	defer m.Unlock()
	now := time.Now().UnixMilli()
	if t := m.persistent.GetTask(result.Task.TaskID); t != nil {
		if t.taskType == string(task.TaskTypeActivity) && *t.lockedBy == m.thisInstanceID {
			m.persistent.RemoveTask(result.Task.TaskID)
			activityCompleted := &history.HistoryEvent{
				Timestamp: now,
				ActivityCompleted: &history.ActivityCompleted{
					TaskScheduledID: result.Task.TaskScheduleEvent.TaskScheduledID,
					ExecutionResult: *result.ExecutionResult,
				},
			}
			payload := m.ForceMarshalHistoryEvent(activityCompleted)
			m.persistent.InsertEvent(&MockDbEvent{
				sequenceNo: m.persistent.NextSeqNo(),
				workflowID: result.Task.WorkflowID,
				createdAt:  now,
				visibleAt:  now,
				payload:    payload,
			})
			return nil
		}
	}
	return errors.New("unexpected error")
}

func (m *mockBackend) AbandonActivityTask(_ context.Context, t *task.ActivityTask, _ *string) error {
	m.Lock()
	defer m.Unlock()
	// unlock task
	if t := m.persistent.GetTask(t.TaskID); t != nil {
		if t.taskType == string(task.TaskTypeActivity) && *t.lockedBy == m.thisInstanceID {
			t.lockedBy = nil
		}
	}
	return nil
}

func (m *mockBackend) ForceUnmarshalHistoryEvent(payload []byte) *history.HistoryEvent {
	he := &history.HistoryEvent{}
	err := m.dataConverter.Unmarshal(payload, he)
	if err != nil {
		panic(err)
	}
	return he
}

func (m *mockBackend) ForceMarshalHistoryEvent(event *history.HistoryEvent) []byte {
	payload, err := m.dataConverter.Marshal(event)
	if err != nil {
		panic(err)
	}
	return payload
}
