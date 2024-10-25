package persistent

import (
	"context"
	"github.com/google/uuid"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql/persistent/base"
	"github.com/tuannh982/simple-workflows-go/pkg/dto/task"
	"gorm.io/gorm"
)

var WorkflowTaskID = uuid.MustParse("00000000-0000-0000-0000-000000000000").String() // reserved id for workflow task

type Task struct {
	WorkflowID    string
	TaskID        string
	TaskType      string
	LockedBy      *string
	LockedAt      int64
	CreatedAt     int64
	VisibleAt     int64
	LastTouch     int64
	Payload       []byte
	ReleaseReason *string
}

type TaskRepository interface {
	InsertTask(ctx context.Context, task *Task) error
	GetTask(ctx context.Context, taskID string) (*Task, error)
	InsertTasks(ctx context.Context, tasks []*Task) error
	ReleaseTask(ctx context.Context, taskID string, taskType task.TaskType, lockedBy string, reason *string) error
	DeleteTask(ctx context.Context, taskID string) error
	TouchTask(ctx context.Context, taskID string) error
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
	//TODO implement me
	panic("implement me")
}

func (r *taskRepository) GetTask(ctx context.Context, taskID string) (*Task, error) {
	//TODO implement me
	panic("implement me")
}

func (r *taskRepository) InsertTasks(ctx context.Context, task []*Task) error {
	//TODO implement me
	panic("implement me")
}

func (r *taskRepository) ReleaseTask(ctx context.Context, taskID string, taskType task.TaskType, lockedBy string, reason *string) error {
	//TODO implement me
	panic("implement me")
}

func (r *taskRepository) DeleteTask(ctx context.Context, taskID string) error {
	//TODO implement me
	panic("implement me")
}

func (r *taskRepository) TouchTask(ctx context.Context, taskID string) error {
	//TODO implement me
	panic("implement me")
}

func (r *taskRepository) GetAndLockAvailableTask(ctx context.Context, taskType task.TaskType, lockedBy string) (*Task, error) {
	//TODO implement me
	panic("implement me")
}

func (r *taskRepository) ResetTaskLastTouchTimestamp(ctx context.Context, workflowID string, taskID string) error {
	//TODO implement me
	panic("implement me")
}
