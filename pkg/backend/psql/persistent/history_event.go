package persistent

import (
	"context"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/psql/persistent/base"
	"gorm.io/gorm"
)

type HistoryEvent struct {
	WorkflowID     string `gorm:"column:workflow_id;type:varchar(255);primaryKey"`
	EventID        string `gorm:"column:event_id;type:uuid;primaryKey"`
	EventTimestamp int64  `gorm:"column:event_timestamp;type:bigint"`
	Payload        []byte `gorm:"column:payload;type:bytea"`
}

type HistoryEventRepository interface {
	InsertHistoryEvents(ctx context.Context, events []*HistoryEvent) error
	GetWorkflowHistory(ctx context.Context, workflowID string) ([]*HistoryEvent, error)
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
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&HistoryEvent{}).CreateInBatches(events, 500)
	return result.Error
}

func (r *historyEventRepository) GetWorkflowHistory(ctx context.Context, workflowID string) ([]*HistoryEvent, error) {
	uow := r.UnitOfWork(ctx)
	var historyEvents []*HistoryEvent
	result := uow.Tx.Model(&HistoryEvent{}).Where("workflow_id = ?", workflowID).Order("event_id").Find(&historyEvents)
	return historyEvents, result.Error
}
