package event

import (
	"context"
	"time"
)

type Event struct {
	ID          string
	Title       string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	CreatorID   string
	Items       []*EventItem
}

type EventItem struct {
	ID              string
	EventID         string
	Name            string
	Price           int64
	MaxParticipants int32
}

type Repository interface {
	Create(ctx context.Context, e *Event) error
	GetByID(ctx context.Context, id string) (*Event, error)
	List(ctx context.Context) ([]*Event, error)
}
