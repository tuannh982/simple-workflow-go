package psql

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/tuannh982/simple-workflows-go/pkg/backend"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql/persistent"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql/persistent/base"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql/persistent/uow"
	"github.com/tuannh982/simple-workflows-go/pkg/dataconverter"
	"github.com/tuannh982/simple-workflows-go/pkg/dto"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/history"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/ptr"
	"github.com/tuannh982/simple-workflows-go/pkg/utils/worker"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"sync"
	"time"
)

type be struct {
	lockedBy         string
	dataConverter    dataconverter.DataConverter
	db               *gorm.DB
	workflowRepo     persistent.WorkflowRepository
	historyEventRepo persistent.HistoryEventRepository
	taskRepo         persistent.TaskRepository
	eventRepo        persistent.EventRepository
	logger           *zap.Logger
	*sync.Mutex
}

func NewPSQLBackend(
	lockedBy string,
	dataConverter dataconverter.DataConverter,
	db *gorm.DB,
	logger *zap.Logger,
) backend.Backend {
	workflowRepo := persistent.NewWorkflowRepository(db)
	historyEventRepo := persistent.NewHistoryEventRepository(db)
	taskRepo := persistent.NewTaskRepository(db)
	eventRepo := persistent.NewEventRepository(db)
	return &be{
		lockedBy:         lockedBy,
		dataConverter:    dataConverter,
		db:               db,
		workflowRepo:     workflowRepo,
		historyEventRepo: historyEventRepo,
		taskRepo:         taskRepo,
		eventRepo:        eventRepo,
		logger:           logger,
		Mutex:            &sync.Mutex{},
	}
}

func (b *be) DataConverter() dataconverter.DataConverter {
	return b.dataConverter
}

func (b *be) getCurrentTimestamp(tx *gorm.DB) int64 {
	type tsHolder struct{ timestamp int64 }
	ts := &tsHolder{}
	tx.Raw("SELECT CAST(EXTRACT(EPOCH FROM NOW()::timestamp) * 1000 AS BIGINT) timestamp;").Scan(ts)
	return ts.timestamp
}

func (b *be) createUow(ctx context.Context, tx *gorm.DB) (context.Context, error) {
	result := tx.Exec(fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", base.IsolationLevelSerializable))
	if result.Error != nil {
		return nil, result.Error
	}
	unitOfWork := uow.NewUnitOfWork(tx)
	uowCtx := unitOfWork.InjectCtx(ctx)
	return uowCtx, nil
}

func (b *be) getCurrentTimestampLocal() int64 {
	return time.Now().UnixMilli()
}

func (b *be) newUuidString() string {
	return uuid.Must(uuid.NewV7()).String()
}

func (b *be) CreateWorkflow(ctx context.Context, info *history.WorkflowExecutionStarted) error {
	err := b.db.Transaction(func(tx *gorm.DB) error {
		uowCtx, err := b.createUow(ctx, tx)
		if err != nil {
			return err
		}
		currentTimestampUTC := b.getCurrentTimestampLocal()
		var parentWorkflowID string
		if info.ParentWorkflowInfo != nil {
			parentWorkflowID = info.ParentWorkflowInfo.WorkflowID
		}
		workflow := persistent.Workflow{
			ID:                   info.WorkflowID,
			Name:                 info.Name,
			Version:              info.Version,
			CreatedAt:            currentTimestampUTC,
			CurrentRuntimeStatus: string(dto.WorkflowRuntimeStatusPending),
			Input:                info.Input,
			ParentWorkflowID:     &parentWorkflowID,
		}
		workflowTask := persistent.Task{
			WorkflowID: info.WorkflowID,
			TaskID:     persistent.WorkflowTaskID,
			TaskType:   string(task.TaskTypeWorkflow),
			CreatedAt:  currentTimestampUTC,
			VisibleAt:  info.ScheduleToStartTimestamp,
			Payload:    info.Input,
		}
		he := &history.HistoryEvent{
			Timestamp:                currentTimestampUTC,
			WorkflowExecutionStarted: info,
		}
		historyEventBytes, err := b.dataConverter.Marshal(he)
		if err != nil {
			return err
		}
		event := persistent.Event{
			WorkflowID: info.WorkflowID,
			EventID:    b.newUuidString(),
			CreatedAt:  currentTimestampUTC,
			VisibleAt:  info.ScheduleToStartTimestamp,
			Payload:    historyEventBytes,
		}
		if err = b.workflowRepo.InsertWorkflow(uowCtx, &workflow); err != nil {
			return err
		}
		if err = b.taskRepo.InsertTask(uowCtx, &workflowTask); err != nil {
			return err
		}
		if err = b.eventRepo.InsertEvents(uowCtx, []*persistent.Event{&event}); err != nil {
			return err
		}
		return nil
	})
	return HandleSQLError(err)
}

func (b *be) GetWorkflowResult(ctx context.Context, name string, workflowID string) (*dto.WorkflowExecutionResult, error) {
	w, err := b.workflowRepo.GetWorkflow(ctx, workflowID)
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

func (b *be) AppendWorkflowEvent(ctx context.Context, workflowID string, event *history.HistoryEvent) error {
	err := b.db.Transaction(func(tx *gorm.DB) error {
		uowCtx, err := b.createUow(ctx, tx)
		if err != nil {
			return err
		}
		currentTimestampUTC := b.getCurrentTimestampLocal()
		historyEventBytes, err := b.dataConverter.Marshal(event)
		if err != nil {
			return err
		}
		e := persistent.Event{
			WorkflowID: workflowID,
			EventID:    b.newUuidString(),
			CreatedAt:  currentTimestampUTC,
			VisibleAt:  currentTimestampUTC,
			Payload:    historyEventBytes,
		}
		if err = b.eventRepo.InsertEvents(uowCtx, []*persistent.Event{&e}); err != nil {
			return err
		}
		if err = b.taskRepo.ResetTaskLastTouchTimestamp(uowCtx, workflowID, persistent.WorkflowTaskID); err != nil {
			return err
		}
		return nil
	})
	return HandleSQLError(err)
}

func (b *be) GetWorkflowHistory(ctx context.Context, workflowID string) ([]*history.HistoryEvent, error) {
	pHistoryEvents, err := b.historyEventRepo.GetWorkflowHistory(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	historyEvents := make([]*history.HistoryEvent, len(pHistoryEvents))
	for i, event := range pHistoryEvents {
		he := &history.HistoryEvent{}
		err = b.dataConverter.Unmarshal(event.Payload, he)
		if err != nil {
			return nil, err
		}
		historyEvents[i] = he
	}
	return historyEvents, nil
}

func (b *be) GetWorkflowTask(ctx context.Context) (result *task.WorkflowTask, err error) {
	tx := b.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	uowCtx, err := b.createUow(ctx, tx)
	if err != nil {
		return nil, HandleSQLError(err)
	}
	currentTimestampUTC := b.getCurrentTimestampLocal()
	t, err := b.taskRepo.GetAndLockAvailableTask(uowCtx, task.TaskTypeWorkflow, b.lockedBy)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, worker.ErrNoTask
		}
		return nil, HandleSQLError(err)
	}
	pHistoryEvents, err := b.historyEventRepo.GetWorkflowHistory(uowCtx, t.WorkflowID)
	if err != nil {
		return nil, err
	}
	pEvents, err := b.eventRepo.GetAvailableWorkflowEventsAndLock(uowCtx, t.WorkflowID, b.lockedBy)
	if err != nil {
		return nil, err
	}
	historyEvents := make([]*history.HistoryEvent, len(pHistoryEvents))
	events := make([]*history.HistoryEvent, len(pEvents))
	for i, event := range pHistoryEvents {
		he := &history.HistoryEvent{}
		err = b.dataConverter.Unmarshal(event.Payload, he)
		if err != nil {
			return nil, err
		}
		historyEvents[i] = he
	}
	for i, event := range pEvents {
		he := &history.HistoryEvent{}
		err = b.dataConverter.Unmarshal(event.Payload, he)
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

func (b *be) CompleteWorkflowTask(ctx context.Context, result *task.WorkflowTaskResult) error {
	err := b.db.Transaction(func(tx *gorm.DB) error {
		uowCtx, err := b.createUow(ctx, tx)
		if err != nil {
			return err
		}
		currentTimestampUTC := b.getCurrentTimestampLocal()
		// touch to trigger transaction lock
		err = b.taskRepo.TouchTask(uowCtx, result.Task.WorkflowID, result.Task.TaskID)
		if err != nil {
			return err
		}
		//
		t, err := b.taskRepo.GetTask(uowCtx, result.Task.WorkflowID, result.Task.TaskID)
		if err != nil {
			return err
		}
		if t.TaskType != string(task.TaskTypeWorkflow) || t.LockedBy == nil || *t.LockedBy != b.lockedBy {
			return fmt.Errorf("task %s of workflow %s is locked by %v, expect %s", t.TaskID, t.WorkflowID, t.LockedBy, b.lockedBy)
		}
		isCompleted := false
		processedEvents := result.Task.NewEvents
		w, err := b.workflowRepo.GetWorkflow(uowCtx, result.Task.WorkflowID)
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
		err = b.workflowRepo.UpdateWorkflow(uowCtx, w.ID, w)
		if err != nil {
			return err
		}
		// delete processed events, and move them to history
		if _, err = b.eventRepo.DeleteEventsByWorkflowIDAndHeldBy(uowCtx, result.Task.WorkflowID, b.lockedBy); err != nil {
			return err
		}
		lastSeqNo, err := b.historyEventRepo.GetLastHistorySeqNo(uowCtx, result.Task.WorkflowID)
		if err != nil {
			return err
		}
		historyEvents := make([]*persistent.HistoryEvent, len(processedEvents))
		for i, event := range processedEvents {
			lastSeqNo++
			bytes, err := b.dataConverter.Marshal(event)
			if err != nil {
				return err
			}
			historyEvents[i] = &persistent.HistoryEvent{
				WorkflowID: result.Task.WorkflowID,
				SequenceNo: lastSeqNo,
				Payload:    bytes,
			}
		}
		if err = b.historyEventRepo.InsertHistoryEvents(uowCtx, historyEvents); err != nil {
			return err
		}
		// build new events list
		pendingTasks := make([]*persistent.Task, 0)
		pendingEvents := make([]*persistent.Event, 0)
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
			heBytes, err := b.dataConverter.Marshal(he)
			if err != nil {
				return err
			}
			pendingTasks = append(pendingTasks, &persistent.Task{
				WorkflowID: result.Task.WorkflowID,
				TaskID:     taskID,
				TaskType:   string(task.TaskTypeActivity),
				CreatedAt:  currentTimestampUTC,
				VisibleAt:  currentTimestampUTC,
				Payload:    bytes,
			})
			pendingEvents = append(pendingEvents, &persistent.Event{
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
			heTimerCreatedBytes, err := b.dataConverter.Marshal(heTimerCreated)
			if err != nil {
				return err
			}
			pendingEvents = append(pendingEvents, &persistent.Event{
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
			heTimerFiredBytes, err := b.dataConverter.Marshal(heTimerFired)
			if err != nil {
				return err
			}
			pendingEvents = append(pendingEvents, &persistent.Event{
				WorkflowID: result.Task.WorkflowID,
				EventID:    b.newUuidString(),
				CreatedAt:  currentTimestampUTC,
				VisibleAt:  timerCreated.FireAt,
				Payload:    heTimerFiredBytes,
			})
		}
		if result.WorkflowExecutionCompleted != nil {
			he := &history.HistoryEvent{
				Timestamp:                  currentTimestampUTC,
				WorkflowExecutionCompleted: result.WorkflowExecutionCompleted,
			}
			bytes, err := b.dataConverter.Marshal(he)
			if err != nil {
				return err
			}
			pendingEvents = append(pendingEvents, &persistent.Event{
				WorkflowID: result.Task.WorkflowID,
				EventID:    b.newUuidString(),
				CreatedAt:  currentTimestampUTC,
				VisibleAt:  currentTimestampUTC,
				Payload:    bytes,
			})
		}
		if err = b.taskRepo.InsertTasks(uowCtx, pendingTasks); err != nil {
			return err
		}
		if err = b.eventRepo.InsertEvents(uowCtx, pendingEvents); err != nil {
			return err
		}
		if isCompleted {
			if err = b.taskRepo.DeleteTask(uowCtx, result.Task.WorkflowID, result.Task.TaskID); err != nil {
				return err
			}
		} else if shouldNotifyWorkflowTask {
			if err = b.taskRepo.ResetTaskLastTouchTimestamp(uowCtx, result.Task.WorkflowID, persistent.WorkflowTaskID); err != nil {
				return err
			}
		}
		return nil
	})
	return HandleSQLError(err)
}

func (b *be) AbandonWorkflowTask(ctx context.Context, t *task.WorkflowTask, reason *string) error {
	err := b.db.Transaction(func(tx *gorm.DB) error {
		uowCtx, err := b.createUow(ctx, tx)
		if err != nil {
			return err
		}
		return b.taskRepo.ReleaseTask(uowCtx, t.WorkflowID, t.TaskID, task.TaskTypeWorkflow, b.lockedBy, reason)
	})
	return HandleSQLError(err)
}

func (b *be) GetActivityTask(ctx context.Context) (result *task.ActivityTask, err error) {
	tx := b.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	uowCtx, err := b.createUow(ctx, tx)
	if err != nil {
		return nil, HandleSQLError(err)
	}
	t, err := b.taskRepo.GetAndLockAvailableTask(uowCtx, task.TaskTypeActivity, b.lockedBy)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, worker.ErrNoTask
		}
		return nil, HandleSQLError(err)
	}
	payload := t.Payload
	activityScheduled := &history.ActivityScheduled{}
	err = b.dataConverter.Unmarshal(payload, activityScheduled)
	if err != nil {
		return nil, HandleSQLError(err)
	}
	return &task.ActivityTask{
		TaskID:            t.TaskID,
		WorkflowID:        t.WorkflowID,
		TaskScheduleEvent: activityScheduled,
	}, nil
}

func (b *be) CompleteActivityTask(ctx context.Context, result *task.ActivityTaskResult) error {
	err := b.db.Transaction(func(tx *gorm.DB) error {
		uowCtx, err := b.createUow(ctx, tx)
		if err != nil {
			return err
		}
		currentTimestampUTC := b.getCurrentTimestampLocal()
		// touch to trigger transaction lock
		err = b.taskRepo.TouchTask(uowCtx, result.Task.WorkflowID, result.Task.TaskID)
		if err != nil {
			return err
		}
		//
		err = b.taskRepo.DeleteTask(uowCtx, result.Task.WorkflowID, result.Task.TaskID)
		if err != nil {
			return err
		}
		if result.ExecutionResult == nil {
			return errors.New("execution result is nil")
		}
		he := &history.HistoryEvent{
			Timestamp: currentTimestampUTC,
			ActivityCompleted: &history.ActivityCompleted{
				TaskScheduledID: result.Task.TaskScheduleEvent.TaskScheduledID,
				ExecutionResult: *result.ExecutionResult,
			},
		}
		bytes, err := b.dataConverter.Marshal(he)
		if err != nil {
			return err
		}
		event := persistent.Event{
			WorkflowID: result.Task.WorkflowID,
			EventID:    b.newUuidString(),
			CreatedAt:  currentTimestampUTC,
			VisibleAt:  currentTimestampUTC,
			Payload:    bytes,
		}
		if err = b.eventRepo.InsertEvents(uowCtx, []*persistent.Event{&event}); err != nil {
			return err
		}
		if err = b.taskRepo.ResetTaskLastTouchTimestamp(uowCtx, result.Task.WorkflowID, persistent.WorkflowTaskID); err != nil {
			return err
		}
		return nil
	})
	return HandleSQLError(err)
}

func (b *be) AbandonActivityTask(ctx context.Context, t *task.ActivityTask, reason *string) error {
	err := b.db.Transaction(func(tx *gorm.DB) error {
		uowCtx, err := b.createUow(ctx, tx)
		if err != nil {
			return err
		}
		return b.taskRepo.ReleaseTask(uowCtx, t.WorkflowID, t.TaskID, task.TaskTypeActivity, b.lockedBy, reason)
	})
	return HandleSQLError(err)
}
