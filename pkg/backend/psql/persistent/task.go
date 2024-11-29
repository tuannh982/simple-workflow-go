package persistent

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/psql/persistent/base"
	"github.com/tuannh982/simple-workflow-go/pkg/dto/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrTaskLost = errors.New("task lost")

var WorkflowTaskID = uuid.MustParse("00000000-0000-0000-0000-000000000000").String() // reserved id for workflow task

type Task struct {
	WorkflowID    string  `gorm:"column:workflow_id;type:varchar(255);primaryKey"`
	TaskID        string  `gorm:"column:task_id;type:uuid;primaryKey"`
	TaskType      string  `gorm:"column:task_type;type:varchar(255);index:idx_task_type_locked_by_visible_at"`
	NumAttempted  int32   `gorm:"column:num_attempted;type:integer"`
	LockedBy      *string `gorm:"column:locked_by;type:varchar(255);index:idx_task_type_locked_by_visible_at"`
	LockedAt      int64   `gorm:"column:locked_at;type:bigint"`
	CreatedAt     int64   `gorm:"column:created_at;type:bigint"`
	VisibleAt     int64   `gorm:"column:visible_at;type:bigint;index:idx_task_type_locked_by_visible_at"`
	LastTouch     int64   `gorm:"column:last_touch;type:bigint"`
	Payload       []byte  `gorm:"column:payload;type:bytea"`
	ReleaseReason *string `gorm:"column:release_reason;type:text"`
}

type TaskRepository interface {
	InsertTask(ctx context.Context, task *Task) error
	GetTask(ctx context.Context, workflowID string, taskID string) (*Task, error)
	InsertTasks(ctx context.Context, tasks []*Task) error
	ReleaseTask(ctx context.Context, workflowID string, taskID string, taskType task.TaskType, lockedBy string, reason *string, nextScheduleTimestamp *int64) error
	DeleteTask(ctx context.Context, workflowID string, taskID string, taskType task.TaskType, lockedBy string) error
	DeleteTaskUnsafe(ctx context.Context, workflowID string, taskID string, taskType task.TaskType) error
	GetAndLockAvailableTask(ctx context.Context, taskType task.TaskType, lockedBy string, lockExpirationDuration time.Duration) (*Task, *string, error)
	ResetTaskLastTouchTimestamp(ctx context.Context, workflowID string, taskID string) error
}

type taskRepository struct {
	base.BaseRepository
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{
		BaseRepository: base.BaseRepository{DB: db},
	}
}

func (r *taskRepository) InsertTask(ctx context.Context, task *Task) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&Task{}).Create(task)
	return result.Error
}

func (r *taskRepository) GetTask(ctx context.Context, workflowID string, taskID string) (*Task, error) {
	uow := r.UnitOfWork(ctx)
	t := &Task{}
	result := uow.Tx.Model(&Task{}).Where("workflow_id = ? AND task_id = ?", workflowID, taskID).First(&t)
	return t, result.Error
}

func (r *taskRepository) InsertTasks(ctx context.Context, task []*Task) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&Task{}).CreateInBatches(task, 500)
	return result.Error
}

func (r *taskRepository) ReleaseTask(ctx context.Context, workflowID string, taskID string, taskType task.TaskType, lockedBy string, reason *string, nextScheduleTimestamp *int64) error {
	uow := r.UnitOfWork(ctx)
	updates := map[string]interface{}{
		"last_touch":     time.Now().UnixMilli(),
		"num_attempted":  gorm.Expr("num_attempted + 1"),
		"release_reason": reason,
		"locked_by":      nil,
	}
	if nextScheduleTimestamp != nil {
		updates["visible_at"] = *nextScheduleTimestamp
	}
	result := uow.Tx.Model(&Task{}).Where(
		"workflow_id = ? AND task_id = ? AND task_type = ? AND locked_by = ?",
		workflowID, taskID, string(taskType), lockedBy,
	).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskLost
	}
	return nil
}

func (r *taskRepository) DeleteTask(ctx context.Context, workflowID string, taskID string, taskType task.TaskType, lockedBy string) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&Task{}).Where(
		"workflow_id = ? AND task_id = ? AND task_type = ? AND locked_by = ?",
		workflowID, taskID, string(taskType), lockedBy,
	).Delete(&Task{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskLost
	}
	return nil
}

func (r *taskRepository) DeleteTaskUnsafe(ctx context.Context, workflowID string, taskID string, taskType task.TaskType) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&Task{}).Where(
		"workflow_id = ? AND task_id = ? AND task_type = ?",
		workflowID, taskID, string(taskType),
	).Delete(&Task{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskLost
	}
	return nil
}

func (r *taskRepository) GetAndLockAvailableTask(ctx context.Context, taskType task.TaskType, lockedBy string, lockExpirationDuration time.Duration) (*Task, *string, error) {
	uow := r.UnitOfWork(ctx)
	now := time.Now().UnixMilli()
	t := &Task{}
	result := uow.Tx.
		Model(&Task{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			"task_type = ? AND (locked_by IS NULL OR locked_at < ?) AND visible_at < ?",
			string(taskType), now-lockExpirationDuration.Milliseconds(), now,
		).
		Order("last_touch ASC").First(&t)
	if result.Error != nil {
		return nil, nil, result.Error
	}
	previousLockedBy := t.LockedBy
	result = uow.Tx.
		Model(&Task{}).
		Where("workflow_id = ? AND task_id = ?", t.WorkflowID, t.TaskID).
		Updates(map[string]interface{}{
			"locked_by":  lockedBy,
			"locked_at":  now,
			"last_touch": now,
		}).First(&t)
	if result.Error != nil {
		return nil, nil, result.Error
	}
	return t, previousLockedBy, nil
}

func (r *taskRepository) ResetTaskLastTouchTimestamp(ctx context.Context, workflowID string, taskID string) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&Task{}).Where("workflow_id = ? AND task_id = ?", workflowID, taskID).Updates(map[string]interface{}{
		"last_touch": 0,
	})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
