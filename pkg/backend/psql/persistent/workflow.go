package persistent

import (
	"context"
	"errors"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql/persistent/base"
	"gorm.io/gorm"
)

var ErrWorkflowNotFound = errors.New("workflow not found")

type Workflow struct {
	ID                   string  `gorm:"column:id;type:varchar(255);primaryKey"`
	Name                 string  `gorm:"column:name;type:varchar(255)"`
	Version              string  `gorm:"column:version;type:varchar(255)"`
	CreatedAt            int64   `gorm:"column:created_at;type:bigint"`
	StartAt              *int64  `gorm:"column:start_at;type:bigint"`
	CompletedAt          *int64  `gorm:"column:completed_at;type:bigint"`
	CurrentRuntimeStatus string  `gorm:"column:current_runtime_status;type:varchar(255)"`
	Input                []byte  `gorm:"column:input;type:bytea"`
	ResultOutput         *[]byte `gorm:"column:result_output;type:bytea"`
	ResultError          *string `gorm:"column:result_error;type:text"`
	ParentWorkflowID     *string `gorm:"column:parent_workflow_id;type:varchar(255)"`
}

type WorkflowRepository interface {
	InsertWorkflow(ctx context.Context, workflow *Workflow) error
	GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error)
	UpdateWorkflow(ctx context.Context, workflowID string, workflow *Workflow) error
}

type workflowRepository struct {
	base.BaseRepository
}

func NewWorkflowRepository(db *gorm.DB) WorkflowRepository {
	return &workflowRepository{
		BaseRepository: base.BaseRepository{DB: db},
	}
}

func (r *workflowRepository) InsertWorkflow(ctx context.Context, workflow *Workflow) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&Workflow{}).Create(workflow)
	return result.Error
}

func (r *workflowRepository) GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error) {
	uow := r.UnitOfWork(ctx)
	workflow := &Workflow{}
	result := uow.Tx.Model(&Workflow{}).Where("id = ?", workflowID).First(workflow)
	if result.Error != nil {
		return nil, result.Error
	}
	return workflow, nil
}

func (r *workflowRepository) UpdateWorkflow(ctx context.Context, workflowID string, workflow *Workflow) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&Workflow{}).Where("id = ?", workflowID).Updates(workflow)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrWorkflowNotFound
	}
	return nil
}
