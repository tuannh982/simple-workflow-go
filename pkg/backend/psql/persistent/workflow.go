package persistent

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql/persistent/base"
	"gorm.io/gorm"
)

type Workflow struct {
	ID                   string  `gorm:"column:id"`
	Name                 string  `gorm:"column:name"`
	Version              string  `gorm:"column:version"`
	CreatedAt            int64   `gorm:"column:created_at"`
	StartAt              *int64  `gorm:"column:start_at"`
	CompletedAt          *int64  `gorm:"column:completed_at"`
	CurrentRuntimeStatus string  `gorm:"column:current_runtime_status"`
	Input                []byte  `gorm:"column:input"`
	ResultOutput         *[]byte `gorm:"column:result_output"`
	ResultError          *string `gorm:"column:result_error"`
	ParentWorkflowID     *string `gorm:"column:parent_workflow_id"`
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
	result := uow.Tx.Create(workflow)
	return result.Error
}

func (r *workflowRepository) GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error) {
	uow := r.UnitOfWork(ctx)
	var workflow *Workflow
	result := uow.Tx.Where("id = ?", workflowID).First(workflow)
	if result.Error != nil {
		return nil, result.Error
	}
	return workflow, nil
}

func (r *workflowRepository) UpdateWorkflow(ctx context.Context, workflowID string, workflow *Workflow) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Save(workflow)
	return result.Error
}
