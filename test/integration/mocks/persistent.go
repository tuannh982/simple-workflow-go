package mocks

import (
	"fmt"
	"sort"
)

// represent DB tables

type db_workflow struct {
	id                   string
	createdAt            int64
	startAt              *int64
	completedAt          *int64
	lastUpdatedAt        *int64
	currentRuntimeStatus *string
	input                []byte
	resultOutput         *[]byte
	resultError          *string
	parentWorkflowID     *string
	// pkey (id)
}

type db_history_event struct {
	workflowID string
	sequenceNo int32
	payload    []byte
	// pkey (workflowID, sequenceNo)
}

type db_task struct {
	sequenceNo int32
	workflowID string
	taskType   string
	lockedBy   *string
	createdAt  int64
	visibleAt  int64
	payload    []byte
	// pkey (sequenceNo auto inc)
}

type db_event struct {
	sequenceNo int32
	workflowID string
	lockedBy   *string
	createdAt  int64
	visibleAt  int64
	payload    []byte
	// pkey (sequenceNo auto inc)
}

func eventsKeys(sequenceNo int32) string {
	return fmt.Sprintf("%d", sequenceNo)
}

type persistent struct {
	workflows_pk      map[string]*db_workflow
	history_events_pk map[string]*db_history_event
	tasks_pk          map[string]*db_task
	events_pk         map[string]*db_event
	seqNo             int32
}

func NewPersistent() *persistent {
	return &persistent{
		workflows_pk:      make(map[string]*db_workflow),
		history_events_pk: make(map[string]*db_history_event),
		tasks_pk:          make(map[string]*db_task),
		events_pk:         make(map[string]*db_event),
		seqNo:             1,
	}
}

func (r *persistent) NextSeqNo() int32 {
	result := r.seqNo
	r.seqNo++
	return result
}

func (r *persistent) InsertWorkflow(workflow *db_workflow) {
	if r.workflows_pk[workflow.id] != nil {
		panic(fmt.Sprintf("Duplicate workflow id %s", workflow.id))
	}
	r.workflows_pk[workflow.id] = workflow
}

func (r *persistent) InsertHistoryEvent(event *db_history_event) {
	key := fmt.Sprintf("%s_%d", event.workflowID, event.sequenceNo)
	if r.history_events_pk[key] != nil {
		panic(fmt.Sprintf("Duplicate history event id %s", key))
	}
	r.history_events_pk[key] = event
}

func (r *persistent) InsertTask(task *db_task) {
	key := fmt.Sprintf("%d", task.sequenceNo)
	if r.tasks_pk[key] != nil {
		panic(fmt.Sprintf("Duplicate task id %s", key))
	}
	r.tasks_pk[key] = task
}

func (r *persistent) InsertEvent(event *db_event) {
	key := fmt.Sprintf("%d", event.sequenceNo)
	if r.events_pk[key] != nil {
		panic(fmt.Sprintf("Duplicate event id %s", key))
	}
	r.events_pk[key] = event
}

func (r *persistent) DeleteEventsByWorkflowAndLock(workflowID string, lockedBy string) {
	for k, v := range r.events_pk {
		if v.workflowID == workflowID && v.lockedBy != nil && *v.lockedBy == lockedBy {
			delete(r.events_pk, k)
		}
	}
}

func (r *persistent) GetWorkflow(workflowID string) *db_workflow {
	return r.workflows_pk[workflowID]
}

func (r *persistent) FilterWorkflows(filter func(*db_workflow) bool) []*db_workflow {
	result := make([]*db_workflow, 0)
	for _, workflow := range r.workflows_pk {
		if filter(workflow) {
			result = append(result, workflow)
		}
	}
	return result
}

func (r *persistent) GetWorkflowHistory(workflowID string) []*db_history_event {
	result := make([]*db_history_event, 0)
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

func (r *persistent) GetTask(seqNo int32) *db_task {
	key := fmt.Sprintf("%d", seqNo)
	return r.tasks_pk[key]
}

func (r *persistent) RemoveTask(seqNo int32) {
	key := fmt.Sprintf("%d", seqNo)
	delete(r.tasks_pk, key)
}

func (r *persistent) FilterTasks(filter func(*db_task) bool) []*db_task {
	result := make([]*db_task, 0)
	for _, task := range r.tasks_pk {
		if filter(task) {
			result = append(result, task)
		}
	}
	return result
}

func (r *persistent) GetWorkflowEvents(workflowID string) []*db_event {
	result := make([]*db_event, 0)
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

func (r *persistent) FilterWorkflowEvents(filter func(*db_event) bool) []*db_event {
	result := make([]*db_event, 0)
	for _, event := range r.events_pk {
		if filter(event) {
			result = append(result, event)
		}
	}
	return result
}
