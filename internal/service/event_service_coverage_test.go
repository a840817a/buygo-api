package service

import (
	"context"
	"testing"
	"time"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/memory"
	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventService_GetEvent(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)
	ctx := context.Background()

	// Prepare data
	e := &event.Event{
		ID:        "evt-1",
		Title:     "Test Event",
		Status:    event.EventStatusActive,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
	}
	require.NoError(t, repo.Create(ctx, e))

	t.Run("Success", func(t *testing.T) {
		got, err := svc.GetEvent(ctx, "evt-1")
		require.NoError(t, err)
		assert.Equal(t, "Test Event", got.Title)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := svc.GetEvent(ctx, "nonexistent")
		assert.Error(t, err)
	})
}
