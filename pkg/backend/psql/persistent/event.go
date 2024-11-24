package persistent

import (
	"context"
	"github.com/tuannh982/simple-workflow-go/pkg/backend/psql/persistent/base"
	"gorm.io/gorm"
	"time"
)

type Event struct {
	WorkflowID string  `gorm:"column:workflow_id;type:varchar(255);primaryKey;index:idx_workflow_id_held_by_visible_at"`
	EventID    string  `gorm:"column:event_id;type:uuid;primaryKey"`
	HeldBy     *string `gorm:"column:held_by;type:varchar(255);index:idx_workflow_id_held_by_visible_at"`
	CreatedAt  int64   `gorm:"column:created_at;type:bigint"`
	VisibleAt  int64   `gorm:"column:visible_at;type:bigint;index:idx_workflow_id_held_by_visible_at"`
	Payload    []byte  `gorm:"column:payload;type:bytea"`
}

type EventRepository interface {
	InsertEvents(ctx context.Context, events []*Event) error
	DeleteEventsByWorkflowID(ctx context.Context, workflowID string) (int64, error)
	DeleteEventsByWorkflowIDAndHeldBy(ctx context.Context, workflowID string, heldBy string) (int64, error)
	GetAvailableWorkflowEventsAndLock(ctx context.Context, workflowID string, heldBy string) ([]*Event, error)
}

type eventRepository struct {
	base.BaseRepository
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{
		BaseRepository: base.BaseRepository{DB: db},
	}
}

func (r *eventRepository) InsertEvents(ctx context.Context, events []*Event) error {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&Event{}).CreateInBatches(events, 500)
	return result.Error
}

func (r *eventRepository) DeleteEventsByWorkflowID(ctx context.Context, workflowID string) (int64, error) {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&Event{}).Where("workflow_id = ?", workflowID).Delete(&Event{})
	return result.RowsAffected, result.Error
}

func (r *eventRepository) DeleteEventsByWorkflowIDAndHeldBy(ctx context.Context, workflowID string, heldBy string) (int64, error) {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&Event{}).Where("workflow_id = ? AND held_by = ?", workflowID, heldBy).Delete(&Event{})
	return result.RowsAffected, result.Error
}

func (r *eventRepository) GetAvailableWorkflowEventsAndLock(ctx context.Context, workflowID string, heldBy string) ([]*Event, error) {
	uow := r.UnitOfWork(ctx)
	now := time.Now().UnixMilli()
	result := uow.Tx.Model(&Event{}).Where(
		"workflow_id = ? AND (held_by IS NULL OR held_by = ?) AND visible_at < ?",
		workflowID,
		heldBy,
		now,
	).Clauses().Updates(map[string]interface{}{
		"held_by": heldBy,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	var events []*Event
	result = uow.Tx.Model(&Event{}).Where("workflow_id = ? AND held_by = ?", workflowID, heldBy).Find(&events)
	return events, result.Error
}
