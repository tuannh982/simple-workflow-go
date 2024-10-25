package persistent

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql/persistent/base"
	"gorm.io/gorm"
)

type Workflow struct {
	ID                   string
	Name                 string
	Version              string
	CreatedAt            int64
	StartAt              *int64
	CompletedAt          *int64
	CurrentRuntimeStatus string
	Input                []byte
	ResultOutput         *[]byte
	ResultError          *string
	ParentWorkflowID     *string
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
	//TODO implement me
	panic("implement me")
}

func (r *workflowRepository) GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error) {
	//TODO implement me
	panic("implement me")
}

func (r *workflowRepository) UpdateWorkflow(ctx context.Context, workflowID string, workflow *Workflow) error {
	//TODO implement me
	panic("implement me")
}
