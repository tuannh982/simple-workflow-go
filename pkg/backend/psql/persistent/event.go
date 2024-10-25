package persistent

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql/persistent/base"
	"gorm.io/gorm"
)

type Event struct {
	WorkflowID string  `gorm:"column:workflow_id"`
	EventID    string  `gorm:"column:event_id"`
	HeldBy     *string `gorm:"column:held_by"`
	CreatedAt  int64   `gorm:"column:created_at"`
	VisibleAt  int64   `gorm:"column:visible_at"`
	Payload    []byte  `gorm:"column:payload"`
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
	result := uow.Tx.CreateInBatches(events, 500)
	return result.Error
}

func (r *eventRepository) DeleteEventsByWorkflowID(ctx context.Context, workflowID string) (int64, error) {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Where("workflow_id = ?", workflowID).Delete(&Event{})
	return result.RowsAffected, result.Error
}

func (r *eventRepository) DeleteEventsByWorkflowIDAndHeldBy(ctx context.Context, workflowID string, heldBy string) (int64, error) {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Where("workflow_id = ? AND held_by = ?", workflowID, heldBy).Delete(&Event{})
	return result.RowsAffected, result.Error
}

func (r *eventRepository) GetAvailableWorkflowEventsAndLock(ctx context.Context, workflowID string, heldBy string) ([]*Event, error) {
	uow := r.UnitOfWork(ctx)
	result := uow.Tx.Model(&Event{}).Where(
		"workflow_id = ? AND (held_by IS NULL OR held_by = ?)",
		workflowID,
		heldBy,
	).Clauses().Updates(map[string]interface{}{
		"held_by": heldBy,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	var events []*Event
	result = uow.Tx.Where("workflow_id = ? AND held_by = ?", workflowID, heldBy).Find(&events)
	return events, result.Error
}
