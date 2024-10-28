package mocks

import (
	"fmt"
	"sort"
)

// represent DB tables

type MockDbWorkflow struct {
	id                   string
	name                 string
	version              string
	createdAt            int64
	startAt              *int64
	completedAt          *int64
	currentRuntimeStatus string
	input                []byte
	resultOutput         *[]byte
	resultError          *string
	parentWorkflowID     *string
	// pkey (id)
}

type MockDbHistoryEvent struct {
	workflowID string
	sequenceNo int64
	payload    []byte
	// pkey (workflowID, sequenceNo)
}

type MockDbTask struct {
	sequenceNo int64
	workflowID string
	taskType   string
	lockedBy   *string
	createdAt  int64
	visibleAt  int64
	payload    []byte
	// pkey (workflowID, sequenceNo)
}

type MockDbEvent struct {
	sequenceNo int64
	workflowID string
	lockedBy   *string
	createdAt  int64
	visibleAt  int64
	payload    []byte
	// pkey (workflowID, sequenceNo)
}

type MockPersistent struct {
	workflows_pk      map[string]*MockDbWorkflow
	history_events_pk map[string]*MockDbHistoryEvent
	tasks_pk          map[string]*MockDbTask
	events_pk         map[string]*MockDbEvent
	seqNo             int64
}

func NewMockPersistent() *MockPersistent {
	return &MockPersistent{
		workflows_pk:      make(map[string]*MockDbWorkflow),
		history_events_pk: make(map[string]*MockDbHistoryEvent),
		tasks_pk:          make(map[string]*MockDbTask),
		events_pk:         make(map[string]*MockDbEvent),
		seqNo:             1,
	}
}

func (r *MockPersistent) NextSeqNo() int64 {
	result := r.seqNo
	r.seqNo++
	return result
}

func (r *MockPersistent) InsertWorkflow(workflow *MockDbWorkflow) {
	if r.workflows_pk[workflow.id] != nil {
		panic(fmt.Sprintf("Duplicate workflow id %s", workflow.id))
	}
	r.workflows_pk[workflow.id] = workflow
}

func (r *MockPersistent) InsertHistoryEvent(event *MockDbHistoryEvent) {
	key := fmt.Sprintf("%s_%d", event.workflowID, event.sequenceNo)
	if r.history_events_pk[key] != nil {
		panic(fmt.Sprintf("Duplicate history event id %s", key))
	}
	r.history_events_pk[key] = event
}

func (r *MockPersistent) InsertTask(task *MockDbTask) {
	key := fmt.Sprintf("%d", task.sequenceNo)
	if r.tasks_pk[key] != nil {
		panic(fmt.Sprintf("Duplicate task id %s", key))
	}
	r.tasks_pk[key] = task
}

func (r *MockPersistent) InsertEvent(event *MockDbEvent) {
	key := fmt.Sprintf("%d", event.sequenceNo)
	if r.events_pk[key] != nil {
		panic(fmt.Sprintf("Duplicate event id %s", key))
	}
	r.events_pk[key] = event
}

func (r *MockPersistent) DeleteEventsByWorkflowAndLock(workflowID string, lockedBy string) {
	for k, v := range r.events_pk {
		if v.workflowID == workflowID && v.lockedBy != nil && *v.lockedBy == lockedBy {
			delete(r.events_pk, k)
		}
	}
}

func (r *MockPersistent) GetWorkflow(workflowID string) *MockDbWorkflow {
	return r.workflows_pk[workflowID]
}

func (r *MockPersistent) FilterWorkflows(filter func(*MockDbWorkflow) bool) []*MockDbWorkflow {
	result := make([]*MockDbWorkflow, 0)
	for _, workflow := range r.workflows_pk {
		if filter(workflow) {
			result = append(result, workflow)
		}
	}
	return result
}

func (r *MockPersistent) GetWorkflowHistory(workflowID string) []*MockDbHistoryEvent {
	result := make([]*MockDbHistoryEvent, 0)
	for _, e := range r.history_events_pk {
		if e.workflowID == workflowID {
			result = append(result, e)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].sequenceNo < result[j].sequenceNo
	})
	return result
}

func (r *MockPersistent) GetTask(taskID string) *MockDbTask {
	key := fmt.Sprintf("%s", taskID)
	return r.tasks_pk[key]
}

func (r *MockPersistent) RemoveTask(taskID string) {
	key := fmt.Sprintf("%s", taskID)
	delete(r.tasks_pk, key)
}

func (r *MockPersistent) FilterTasks(filter func(*MockDbTask) bool) []*MockDbTask {
	result := make([]*MockDbTask, 0)
	for _, task := range r.tasks_pk {
		if filter(task) {
			result = append(result, task)
		}
	}
	return result
}

func (r *MockPersistent) GetWorkflowEvents(workflowID string) []*MockDbEvent {
	result := make([]*MockDbEvent, 0)
	for _, e := range r.events_pk {
		if e.workflowID == workflowID {
			result = append(result, e)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].sequenceNo < result[j].sequenceNo
	})
	return result
}

func (r *MockPersistent) FilterWorkflowEvents(filter func(*MockDbEvent) bool) []*MockDbEvent {
	result := make([]*MockDbEvent, 0)
	for _, event := range r.events_pk {
		if filter(event) {
			result = append(result, event)
		}
	}
	return result
}
