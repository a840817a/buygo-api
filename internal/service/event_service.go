package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"buygo/internal/domain/event"
)

type EventService struct {
	repo event.Repository
}

func NewEventService(repo event.Repository) *EventService {
	return &EventService{
		repo: repo,
	}
}

func (s *EventService) CreateEvent(ctx context.Context, userID string, title string, start, end time.Time) (*event.Event, error) {
	e := &event.Event{
		ID:        uuid.New().String(),
		Title:     title,
		StartTime: start,
		EndTime:   end,
		CreatorID: userID,
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *EventService) ListEvents(ctx context.Context) ([]*event.Event, error) {
	return s.repo.List(ctx)
}
