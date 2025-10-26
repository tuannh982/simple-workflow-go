package backend

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	persistent2 "github.com/tuannh982/simple-workflow-go/pkg/backend/persistent"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/persistent/base"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/persistent/uow"
	"github.com/tuannh982/simple-workflow-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflow-go/pkg/dto"
	"github.com/tuannh982/simple-workflow-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflow-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflow-go/pkg/utils/ptr"
	"github.com/tuannh982/simple-workflow-go/pkg/utils/worker"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Backend interface {
	DataConverter() dataconverter.Codec
	CreateWorkflow(ctx context.Context, info *history.WorkflowExecutionStarted) error
	GetWorkflowResult(ctx context.Context, name string, workflowID string) (*dto.WorkflowExecutionResult, error)
	AppendWorkflowEvent(ctx context.Context, workflowID string, event *history.HistoryEvent) error
	GetWorkflowHistory(ctx context.Context, workflowID string) ([]*history.HistoryEvent, error)
	GetWorkflowTask(ctx context.Context) (*task.WorkflowTask, error)
	CompleteWorkflowTask(ctx context.Context, result *task.WorkflowTaskResult) error
	AbandonWorkflowTask(ctx context.Context, task *task.WorkflowTask, reason *string) error
	GetActivityTask(ctx context.Context) (*task.ActivityTask, error)
	CompleteActivityTask(ctx context.Context, result *task.ActivityTaskResult) error
	AbandonActivityTask(ctx context.Context, task *task.ActivityTask, reason *string, nextExecutionTime time.Time, stateData []byte) error
}

// TODO: There are some SQL statements that might be specific to Postgres only, refactor that to be generic
type SimpleWorkflowGoBackend struct {
	LockedBy               string
	LockExpirationDuration time.Duration
	Codec                  dataconverter.Codec
	DB                     *gorm.DB
	WorkflowRepo           persistent2.WorkflowRepository
	HistoryEventRepo       persistent2.HistoryEventRepository
	TaskRepo               persistent2.TaskRepository
	EventRepo              persistent2.EventRepository
	Logger                 *zap.Logger
	WorkflowTaskMu         *sync.Mutex
	ActivityTaskMu         *sync.Mutex
}

func (b *SimpleWorkflowGoBackend) DataConverter() dataconverter.Codec {
	return b.Codec
}

func (b *SimpleWorkflowGoBackend) getCurrentTimestamp(tx *gorm.DB) int64 {
	type tsHolder struct{ timestamp int64 }
	ts := &tsHolder{}
	tx.Raw("SELECT CAST(EXTRACT(EPOCH FROM NOW()::timestamp) * 1000 AS BIGINT) timestamp;").Scan(ts)
	return ts.timestamp
}

func (b *SimpleWorkflowGoBackend) createUow(ctx context.Context, tx *gorm.DB) (context.Context, error) {
	result := tx.Exec(fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", base.IsolationLevelSerializable))
	if result.Error != nil {
		return nil, result.Error
	}
	unitOfWork := uow.NewUnitOfWork(tx)
	uowCtx := unitOfWork.InjectCtx(ctx)
	return uowCtx, nil
}

func (b *SimpleWorkflowGoBackend) getCurrentTimestampLocal() int64 {
	return time.Now().UnixMilli()
}

func (b *SimpleWorkflowGoBackend) newUuidString() string {
	return uuid.Must(uuid.NewV6()).String()
}

func (b *SimpleWorkflowGoBackend) CreateWorkflow(ctx context.Context, info *history.WorkflowExecutionStarted) error {
	err := b.DB.Transaction(func(tx *gorm.DB) error {
		uowCtx, err := b.createUow(ctx, tx)
		if err != nil {
			return err
		}
		currentTimestampUTC := b.getCurrentTimestampLocal()
		var parentWorkflowID string
		if info.ParentWorkflowInfo != nil {
			parentWorkflowID = info.ParentWorkflowInfo.WorkflowID
		}
		workflow := persistent2.Workflow{
			ID:                   info.WorkflowID,
			Name:                 info.Name,
			Version:              info.Version,
			CreatedAt:            currentTimestampUTC,
			CurrentRuntimeStatus: string(dto.WorkflowRuntimeStatusPending),
			Input:                info.Input,
			ParentWorkflowID:     &parentWorkflowID,
		}
		workflowTask := persistent2.Task{
			WorkflowID: info.WorkflowID,
			TaskID:     persistent2.WorkflowTaskID,
			TaskType:   string(task.TaskTypeWorkflow),
			CreatedAt:  currentTimestampUTC,
			VisibleAt:  info.ScheduleToStartTimestamp,
			Payload:    info.Input,
		}
		he := &history.HistoryEvent{
			Timestamp:                info.ScheduleToStartTimestamp,
			WorkflowExecutionStarted: info,
		}
		historyEventBytes, err := b.Codec.Marshal(he)
		if err != nil {
			return err
		}
		event := persistent2.Event{
			WorkflowID: info.WorkflowID,
			EventID:    b.newUuidString(),
			CreatedAt:  currentTimestampUTC,
			VisibleAt:  info.ScheduleToStartTimestamp,
			Payload:    historyEventBytes,
		}
		if err = b.WorkflowRepo.InsertWorkflow(uowCtx, &workflow); err != nil {
			return err
		}
		if err = b.TaskRepo.InsertTask(uowCtx, &workflowTask); err != nil {
			return err
		}
		if err = b.EventRepo.InsertEvents(uowCtx, []*persistent2.Event{&event}); err != nil {
			return err
		}
		return nil
	})
	return HandleSQLError(err)
}

func (b *SimpleWorkflowGoBackend) GetWorkflowResult(ctx context.Context, name string, workflowID string) (*dto.WorkflowExecutionResult, error) {
	w, err := b.WorkflowRepo.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	if w.Name != name {
		return nil, fmt.Errorf("workflow name %s does not match expected workflow name %s", w.Name, name)
	}
	executionResult := dto.ExecutionResult{
		Result: w.ResultOutput,
	}
	if w.ResultError != nil {
		executionResult.Error = &dto.Error{Message: *w.ResultError}
	}
	return &dto.WorkflowExecutionResult{
		WorkflowID:      w.ID,
		Version:         w.Version,
		RuntimeStatus:   w.CurrentRuntimeStatus,
		ExecutionResult: executionResult,
	}, nil
}

func (b *SimpleWorkflowGoBackend) AppendWorkflowEvent(ctx context.Context, workflowID string, event *history.HistoryEvent) error {
	err := b.DB.Transaction(func(tx *gorm.DB) error {
		uowCtx, err := b.createUow(ctx, tx)
		if err != nil {
			return err
		}
		currentTimestampUTC := b.getCurrentTimestampLocal()
		if event.Timestamp == 0 {
			event.Timestamp = currentTimestampUTC
		}
		historyEventBytes, err := b.Codec.Marshal(event)
		if err != nil {
			return err
		}
		e := persistent2.Event{
			WorkflowID: workflowID,
			EventID:    b.newUuidString(),
			CreatedAt:  currentTimestampUTC,
			VisibleAt:  event.Timestamp,
			Payload:    historyEventBytes,
		}
		if err = b.EventRepo.InsertEvents(uowCtx, []*persistent2.Event{&e}); err != nil {
			return err
		}
		if err = b.TaskRepo.ResetTaskLastTouchTimestamp(uowCtx, workflowID, persistent2.WorkflowTaskID); err != nil {
			return err
		}
		return nil
	})
	return HandleSQLError(err)
}

func (b *SimpleWorkflowGoBackend) GetWorkflowHistory(ctx context.Context, workflowID string) ([]*history.HistoryEvent, error) {
	pHistoryEvents, err := b.HistoryEventRepo.GetWorkflowHistory(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	historyEvents := make([]*history.HistoryEvent, len(pHistoryEvents))
	for i, event := range pHistoryEvents {
		he := &history.HistoryEvent{}
		err = b.Codec.Unmarshal(event.Payload, he)
		if err != nil {
			return nil, err
		}
		historyEvents[i] = he
	}
	return historyEvents, nil
}

func (b *SimpleWorkflowGoBackend) GetWorkflowTask(ctx context.Context) (result *task.WorkflowTask, err error) {
	b.WorkflowTaskMu.Lock()
	defer b.WorkflowTaskMu.Unlock()
	var t *persistent2.Task
	tx := b.DB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
			if t != nil {
				tErr := b.TaskRepo.TouchTask(ctx, t.WorkflowID, t.TaskID)
				if tErr != nil {
					b.Logger.Error(
						"failed to update workflow task",
						zap.Error(tErr),
						zap.String("workflow_id", t.WorkflowID),
					)
				}
			}
		} else {
			tx.Commit()
		}
	}()
	uowCtx, err := b.createUow(ctx, tx)
	if err != nil {
		return nil, HandleSQLError(err)
	}
	currentTimestampUTC := b.getCurrentTimestampLocal()
	t, previouslyLockedBy, err := b.TaskRepo.GetAndLockAvailableTask(uowCtx, task.TaskTypeWorkflow, b.LockedBy, b.LockExpirationDuration)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, worker.ErrNoTask
		}
		return nil, HandleSQLError(err)
	}
	pHistoryEvents, err := b.HistoryEventRepo.GetWorkflowHistory(uowCtx, t.WorkflowID)
	if err != nil {
		return nil, err
	}
	pEvents, err := b.EventRepo.GetAvailableWorkflowEventsAndLock(uowCtx, t.WorkflowID, b.LockedBy, previouslyLockedBy)
	if err != nil {
		return nil, err
	}
	historyEvents := make([]*history.HistoryEvent, len(pHistoryEvents))
	events := make([]*history.HistoryEvent, len(pEvents))
	for i, event := range pHistoryEvents {
		he := &history.HistoryEvent{}
		err = b.Codec.Unmarshal(event.Payload, he)
		if err != nil {
			return nil, err
		}
		historyEvents[i] = he
	}
	for i, event := range pEvents {
		he := &history.HistoryEvent{}
		err = b.Codec.Unmarshal(event.Payload, he)
		if err != nil {
			return nil, err
		}
		events[i] = he
	}
	if len(events) == 0 {
		return nil, worker.ErrNoTask
	}
	return &task.WorkflowTask{
		TaskID:         t.TaskID,
		WorkflowID:     t.WorkflowID,
		FetchTimestamp: currentTimestampUTC,
		OldEvents:      historyEvents,
		NewEvents:      events,
	}, nil
}

func (b *SimpleWorkflowGoBackend) CompleteWorkflowTask(ctx context.Context, result *task.WorkflowTaskResult) error {
	b.WorkflowTaskMu.Lock()
	defer b.WorkflowTaskMu.Unlock()
	err := b.DB.Transaction(func(tx *gorm.DB) error {
		uowCtx, err := b.createUow(ctx, tx)
		if err != nil {
			return err
		}
		currentTimestampUTC := b.getCurrentTimestampLocal()
		if err = b.TaskRepo.ReleaseTask(uowCtx, result.Task.WorkflowID, result.Task.TaskID, task.TaskTypeWorkflow, b.LockedBy, nil, nil, nil); err != nil {
			return err
		}
		isCompleted := false
		processedEvents := result.Task.NewEvents
		w, err := b.WorkflowRepo.GetWorkflow(uowCtx, result.Task.WorkflowID)
		if err != nil {
			return err
		}
		// update workflow state
		for _, event := range processedEvents {
			if event.WorkflowExecutionStarted != nil {
				w.StartAt = &event.Timestamp
				w.CurrentRuntimeStatus = string(dto.WorkflowRuntimeStatusRunning)
			} else if event.WorkflowExecutionCompleted != nil {
				w.CompletedAt = &event.Timestamp
				w.CurrentRuntimeStatus = string(dto.WorkflowRuntimeStatusCompleted)
				w.ResultOutput = event.WorkflowExecutionCompleted.Result
				if event.WorkflowExecutionCompleted.Error != nil {
					w.ResultError = ptr.Ptr(event.WorkflowExecutionCompleted.Error.Message)
				}
				isCompleted = true
			}
		}
		err = b.WorkflowRepo.UpdateWorkflow(uowCtx, w.ID, w)
		if err != nil {
			return err
		}
		// delete processed events, and move them to history
		if _, err = b.EventRepo.DeleteEventsByWorkflowIDAndHeldBy(uowCtx, result.Task.WorkflowID, b.LockedBy); err != nil {
			return err
		}
		historyEvents := make([]*persistent2.HistoryEvent, len(processedEvents))
		for i, event := range processedEvents {
			bytes, err := b.Codec.Marshal(event)
			if err != nil {
				return err
			}
			historyEvents[i] = &persistent2.HistoryEvent{
				WorkflowID:     result.Task.WorkflowID,
				EventID:        b.newUuidString(),
				EventTimestamp: event.Timestamp,
				Payload:        bytes,
			}
		}
		if err = b.HistoryEventRepo.InsertHistoryEvents(uowCtx, historyEvents); err != nil {
			return err
		}
		// build new events list
		pendingTasks := make([]*persistent2.Task, 0)
		pendingEvents := make([]*persistent2.Event, 0)
		shouldNotifyWorkflowTask := len(result.PendingActivities) != 0 || len(result.PendingTimers) != 0
		for _, activityScheduled := range result.PendingActivities {
			bytes, err := dto.Marshal(activityScheduled)
			if err != nil {
				return err
			}
			taskID := b.newUuidString()
			he := &history.HistoryEvent{
				Timestamp:         currentTimestampUTC,
				ActivityScheduled: activityScheduled,
			}
			heBytes, err := b.Codec.Marshal(he)
			if err != nil {
				return err
			}
			pendingTasks = append(pendingTasks, &persistent2.Task{
				WorkflowID: result.Task.WorkflowID,
				TaskID:     taskID,
				TaskType:   string(task.TaskTypeActivity),
				CreatedAt:  currentTimestampUTC,
				VisibleAt:  currentTimestampUTC,
				Payload:    bytes,
			})
			pendingEvents = append(pendingEvents, &persistent2.Event{
				WorkflowID: result.Task.WorkflowID,
				EventID:    taskID,
				CreatedAt:  currentTimestampUTC,
				VisibleAt:  currentTimestampUTC,
				Payload:    heBytes,
			})
		}
		for _, timerCreated := range result.PendingTimers {
			heTimerCreated := &history.HistoryEvent{
				Timestamp:    currentTimestampUTC,
				TimerCreated: timerCreated,
			}
			heTimerCreatedBytes, err := b.Codec.Marshal(heTimerCreated)
			if err != nil {
				return err
			}
			pendingEvents = append(pendingEvents, &persistent2.Event{
				WorkflowID: result.Task.WorkflowID,
				EventID:    b.newUuidString(),
				CreatedAt:  currentTimestampUTC,
				VisibleAt:  currentTimestampUTC,
				Payload:    heTimerCreatedBytes,
			})
			heTimerFired := &history.HistoryEvent{
				Timestamp:  timerCreated.FireAt,
				TimerFired: &history.TimerFired{TimerID: timerCreated.TimerID},
			}
			heTimerFiredBytes, err := b.Codec.Marshal(heTimerFired)
			if err != nil {
				return err
			}
			pendingEvents = append(pendingEvents, &persistent2.Event{
				WorkflowID: result.Task.WorkflowID,
				EventID:    b.newUuidString(),
				CreatedAt:  currentTimestampUTC,
				VisibleAt:  timerCreated.FireAt,
				Payload:    heTimerFiredBytes,
			})
		}
		if result.WorkflowExecutionCompleted != nil {
			if !isCompleted { // WorkflowExecutionCompleted is not in processed event list
				he := &history.HistoryEvent{
					Timestamp:                  currentTimestampUTC,
					WorkflowExecutionCompleted: result.WorkflowExecutionCompleted,
				}
				bytes, err := b.Codec.Marshal(he)
				if err != nil {
					return err
				}
				pendingEvents = append(pendingEvents, &persistent2.Event{
					WorkflowID: result.Task.WorkflowID,
					EventID:    b.newUuidString(),
					CreatedAt:  currentTimestampUTC,
					VisibleAt:  currentTimestampUTC,
					Payload:    bytes,
				})
			}
		}
		if err = b.TaskRepo.InsertTasks(uowCtx, pendingTasks); err != nil {
			return err
		}
		if err = b.EventRepo.InsertEvents(uowCtx, pendingEvents); err != nil {
			return err
		}
		if isCompleted {
			if err = b.TaskRepo.DeleteTaskUnsafe(uowCtx, result.Task.WorkflowID, result.Task.TaskID, task.TaskTypeWorkflow); err != nil {
				return err
			}
		} else if shouldNotifyWorkflowTask {
			if err = b.TaskRepo.ResetTaskLastTouchTimestamp(uowCtx, result.Task.WorkflowID, persistent2.WorkflowTaskID); err != nil {
				return err
			}
		}
		return nil
	})
	return HandleSQLError(err)
}

func (b *SimpleWorkflowGoBackend) AbandonWorkflowTask(ctx context.Context, t *task.WorkflowTask, reason *string) error {
	b.WorkflowTaskMu.Lock()
	defer b.WorkflowTaskMu.Unlock()
	err := b.DB.Transaction(func(tx *gorm.DB) error {
		uowCtx, err := b.createUow(ctx, tx)
		if err != nil {
			return err
		}
		err = b.TaskRepo.ReleaseTask(uowCtx, t.WorkflowID, t.TaskID, task.TaskTypeWorkflow, b.LockedBy, reason, nil, nil)
		if err != nil {
			return err
		}
		_, err = b.EventRepo.ReleaseEventsByWorkflowIDAndHeldBy(uowCtx, t.WorkflowID, b.LockedBy)
		if err != nil {
			return err
		}
		return nil
	})
	return HandleSQLError(err)
}

func (b *SimpleWorkflowGoBackend) GetActivityTask(ctx context.Context) (result *task.ActivityTask, err error) {
	b.ActivityTaskMu.Lock()
	defer b.ActivityTaskMu.Unlock()
	tx := b.DB.Begin()
	var t *persistent2.Task
	defer func() {
		if err != nil {
			tx.Rollback()
			if t != nil {
				tErr := b.TaskRepo.TouchTask(ctx, t.WorkflowID, t.TaskID)
				if tErr != nil {
					b.Logger.Error(
						"failed to update activity task",
						zap.Error(tErr),
						zap.String("workflow_id", t.WorkflowID),
						zap.String("task_id", t.TaskID),
					)
				}
			}
		} else {
			tx.Commit()
		}
	}()
	uowCtx, err := b.createUow(ctx, tx)
	if err != nil {
		return nil, HandleSQLError(err)
	}
	t, _, err = b.TaskRepo.GetAndLockAvailableTask(uowCtx, task.TaskTypeActivity, b.LockedBy, b.LockExpirationDuration)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, worker.ErrNoTask
		}
		return nil, HandleSQLError(err)
	}
	payload := t.Payload
	activityScheduled := &history.ActivityScheduled{}
	err = b.Codec.Unmarshal(payload, activityScheduled)
	if err != nil {
		return nil, HandleSQLError(err)
	}
	return &task.ActivityTask{
		TaskID:            t.TaskID,
		WorkflowID:        t.WorkflowID,
		NumAttempted:      int(t.NumAttempted),
		StateData:         t.StateData,
		TaskScheduleEvent: activityScheduled,
	}, nil
}

func (b *SimpleWorkflowGoBackend) CompleteActivityTask(ctx context.Context, result *task.ActivityTaskResult) error {
	b.ActivityTaskMu.Lock()
	defer b.ActivityTaskMu.Unlock()
	err := b.DB.Transaction(func(tx *gorm.DB) error {
		uowCtx, err := b.createUow(ctx, tx)
		if err != nil {
			return err
		}
		currentTimestampUTC := b.getCurrentTimestampLocal()
		if result.ExecutionResult == nil {
			return errors.New("execution result is nil")
		}
		he := &history.HistoryEvent{
			Timestamp: currentTimestampUTC,
			ActivityCompleted: &history.ActivityCompleted{
				TaskScheduledID:   result.Task.TaskScheduleEvent.TaskScheduledID,
				ExecutionResult:   *result.ExecutionResult,
				ActivityStateData: result.GetStateData(),
			},
		}
		bytes, err := b.Codec.Marshal(he)
		if err != nil {
			return err
		}
		event := persistent2.Event{
			WorkflowID: result.Task.WorkflowID,
			EventID:    b.newUuidString(),
			CreatedAt:  currentTimestampUTC,
			VisibleAt:  currentTimestampUTC,
			Payload:    bytes,
		}
		if err = b.TaskRepo.DeleteTask(uowCtx, result.Task.WorkflowID, result.Task.TaskID, task.TaskTypeActivity, b.LockedBy); err != nil {
			return err
		}
		if err = b.EventRepo.InsertEvents(uowCtx, []*persistent2.Event{&event}); err != nil {
			return err
		}
		if err = b.TaskRepo.ResetTaskLastTouchTimestamp(uowCtx, result.Task.WorkflowID, persistent2.WorkflowTaskID); err != nil {
			return err
		}
		return nil
	})
	return HandleSQLError(err)
}

func (b *SimpleWorkflowGoBackend) AbandonActivityTask(ctx context.Context, t *task.ActivityTask, reason *string, nextExecutionTime time.Time, stateData []byte) error {
	b.ActivityTaskMu.Lock()
	defer b.ActivityTaskMu.Unlock()
	err := b.DB.Transaction(func(tx *gorm.DB) error {
		uowCtx, err := b.createUow(ctx, tx)
		if err != nil {
			return err
		}
		currentTimestampUTC := b.getCurrentTimestampLocal()
		nextExecutionTimeUTC := nextExecutionTime.UnixMilli()
		nextScheduleTimestamp := currentTimestampUTC
		if nextExecutionTimeUTC > currentTimestampUTC {
			nextScheduleTimestamp = nextExecutionTimeUTC
		}
		return b.TaskRepo.ReleaseTask(uowCtx, t.WorkflowID, t.TaskID, task.TaskTypeActivity, b.LockedBy, reason, &nextScheduleTimestamp, stateData)
	})
	return HandleSQLError(err)
}
