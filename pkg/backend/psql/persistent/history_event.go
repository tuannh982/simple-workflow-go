package persistent

import (
	"context"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/psql/persistent/base"
	"gorm.io/gorm"
)

type HistoryEvent struct {
	WorkflowID string `gorm:"column:workflow_id;type:varchar(255);primaryKey"`
	SequenceNo int64  `gorm:"column:sequence_no;type:bigint;primaryKey"`
	Payload    []byte `gorm:"column:payload;type:bytea"`
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
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&HistoryEvent{}).CreateInBatches(events, 500)
	return result.Error
}

func (r *historyEventRepository) GetWorkflowHistory(ctx context.Context, workflowID string) ([]*HistoryEvent, error) {
	uow := r.UnitOfWork(ctx)
	var historyEvents []*HistoryEvent
	result := uow.Tx.Model(&HistoryEvent{}).Where("workflow_id = ?", workflowID).Find(&historyEvents)
	return historyEvents, result.Error
}

func (r *historyEventRepository) GetLastHistorySeqNo(ctx context.Context, workflowID string) (int64, error) {
	uow := r.UnitOfWork(ctx)
	var maxValue int64
	result := uow.Tx.
		Model(&HistoryEvent{}).
		Where("workflow_id = ?", workflowID).
		Select("COALESCE(MAX(sequence_no), 0)").
		Scan(&maxValue)
	if result.Error != nil {
		return 0, result.Error
	}
	return maxValue, nil
}
