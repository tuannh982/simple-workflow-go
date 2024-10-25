package persistent

import (
	"context"
	"github.com/tuannh982/simple-workflows-go/pkg/backend/psql/persistent/base"
	"gorm.io/gorm"
)

type Event struct {
	WorkflowID string
	EventID    string
	LockedBy   *string
	CreatedAt  int64
	VisibleAt  int64
	Payload    []byte
}

type EventRepository interface {
	InsertEvents(ctx context.Context, events []*Event) error
	DeleteEventsByWorkflowID(ctx context.Context, workflowID string) (int, error)
	DeleteEventsByWorkflowIDAndLockedBy(ctx context.Context, workflowID string, lockedBy string) error
	GetAvailableWorkflowEventsAndLock(ctx context.Context, workflowID string, lockedBy string) ([]*Event, error)
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
	//TODO implement me
	panic("implement me")
}

func (r *eventRepository) DeleteEventsByWorkflowID(ctx context.Context, workflowID string) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (r *eventRepository) DeleteEventsByWorkflowIDAndLockedBy(ctx context.Context, workflowID string, lockedBy string) error {
	//TODO implement me
	panic("implement me")
}

func (r *eventRepository) GetAvailableWorkflowEventsAndLock(ctx context.Context, workflowID string, lockedBy string) ([]*Event, error) {
	//TODO implement me
	panic("implement me")
}
