package persistent

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql/persistent/base"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrTaskLost = errors.New("task lost")

var WorkflowTaskID = uuid.MustParse("00000000-0000-0000-0000-000000000000").String() // reserved id for workflow task

type Task struct {
	WorkflowID    string  `gorm:"column:workflow_id"`
	TaskID        string  `gorm:"column:task_id"`
	TaskType      string  `gorm:"column:task_type"`
	LockedBy      *string `gorm:"column:locked_by"`
	LockedAt      int64   `gorm:"column:locked_at"` // TODO implement later
	CreatedAt     int64   `gorm:"column:created_at"`
	VisibleAt     int64   `gorm:"column:visible_at"`
	LastTouch     int64   `gorm:"column:last_touch"`
	Payload       []byte  `gorm:"column:payload"`
	ReleaseReason *string `gorm:"column:release_reason"`
}

type TaskRepository interface {
	InsertTask(ctx context.Context, task *Task) error
	GetTask(ctx context.Context, workflowID string, taskID string) (*Task, error)
	InsertTasks(ctx context.Context, tasks []*Task) error
	ReleaseTask(ctx context.Context, workflowID string, taskID string, taskType task.TaskType, lockedBy string, reason *string) error
	DeleteTask(ctx context.Context, workflowID string, taskID string) error
	TouchTask(ctx context.Context, workflowID string, taskID string) error
	GetAndLockAvailableTask(ctx context.Context, taskType task.TaskType, lockedBy string) (*Task, error)
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
	result := uow.Tx.Create(task)
	return result.Error
}

func (r *taskRepository) GetTask(ctx context.Context, workflowID string, taskID string) (*Task, error) {
	uow := r.UnitOfWork(ctx)
	var t *Task
	result := uow.Tx.Where("workflow_id = ? AND task_id = ?", workflowID, taskID).First(&t)
	return t, result.Error
}

func (r *taskRepository) InsertTasks(ctx context.Context, task []*Task) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.CreateInBatches(task, 500)
	return result.Error
}

func (r *taskRepository) ReleaseTask(ctx context.Context, workflowID string, taskID string, taskType task.TaskType, lockedBy string, reason *string) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Where("workflow_id = ? AND task_id = ? AND task_type = ? AND locked_by = ?", workflowID, taskID, string(taskType), lockedBy).Updates(map[string]interface{}{
		"release_reason": reason,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskLost
	}
	return nil
}

func (r *taskRepository) DeleteTask(ctx context.Context, workflowID string, taskID string) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Where("workflow_id = ? task_id = ?", workflowID, taskID).Delete(&Task{})
	return result.Error
}

func (r *taskRepository) TouchTask(ctx context.Context, workflowID string, taskID string) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Where("workflow_id = ? task_id = ?", workflowID, taskID).Updates(map[string]interface{}{
		"last_touch": time.Now().UnixMilli(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskLost
	}
	return nil
}

func (r *taskRepository) GetAndLockAvailableTask(ctx context.Context, taskType task.TaskType, lockedBy string) (*Task, error) {
	uow := r.UnitOfWork(ctx)
	var t *Task
	result := uow.Tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("task_type = ? AND (locked_by IS NULL OR locked_by = ?)", taskType, lockedBy).
		Order("last_touch ASC").
		First(&t)
	if result.Error != nil {
		return nil, result.Error
	}
	return t, nil
}

func (r *taskRepository) ResetTaskLastTouchTimestamp(ctx context.Context, workflowID string, taskID string) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Where("workflow_id = ? AND task_id = ?", workflowID, taskID).Updates(map[string]interface{}{
		"last_touch": 0,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskLost
	}
	return nil
}
