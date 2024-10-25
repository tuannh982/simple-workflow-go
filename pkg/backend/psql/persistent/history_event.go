package persistent

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql/persistent/base"
	"gorm.io/gorm"
)

type HistoryEvent struct {
	WorkflowID string
	SequenceNo int64
	Payload    []byte
}

type HistoryEventRepository interface {
	InsertHistoryEvents(ctx context.Context, events []*HistoryEvent) error
	GetWorkflowHistory(ctx context.Context, workflowID string) ([]*HistoryEvent, error)
	GetLastHistorySeqNo(ctx context.Context, workflowID string) (int64, error)
}

type historyEventRepository struct {
	base.BaseRepository
}

func NewHistoryEventRepository(db *gorm.DB) HistoryEventRepository {
	return &historyEventRepository{
		BaseRepository: base.BaseRepository{DB: db},
	}
}

func (r *historyEventRepository) InsertHistoryEvents(ctx context.Context, events []*HistoryEvent) error {
	//TODO implement me
	panic("implement me")
}

func (r *historyEventRepository) GetWorkflowHistory(ctx context.Context, workflowID string) ([]*HistoryEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (r *historyEventRepository) GetLastHistorySeqNo(ctx context.Context, workflowID string) (int64, error) {
	//TODO implement me
	panic("implement me")
}
