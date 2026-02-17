package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventRepository_Comprehensive(t *testing.T) {
	db := newEventTestDB(t)
	repo := NewEventRepository(db)
	ctx := context.Background()

	// 1. Create Event
	e := &event.Event{
		ID:        "evt-1",
		Title:     "Event 1",
		Status:    event.EventStatusActive,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(24 * time.Hour),
		Items: []*event.EventItem{
			{ID: "item-1", Name: "Item 1", Price: 100},
		},
	}
	require.NoError(t, repo.Create(ctx, e))

	// 2. List (Public)
	res, err := repo.List(ctx, 10, 0, "", false, false)
	require.NoError(t, err)
	assert.Len(t, res, 1)

	// 3. Register
	reg := &event.Registration{
		ID:      "reg-1",
		EventID: "evt-1",
		UserID:  "user-1",
		Status:  event.RegistrationStatusPending,
		SelectedItems: []*event.RegistrationItem{
			{EventItemID: "item-1", Quantity: 1},
		},
	}
	require.NoError(t, repo.Register(ctx, reg))

	// 4. GetRegistration
	gotReg, err := repo.GetRegistration(ctx, "reg-1")
	require.NoError(t, err)
	assert.Equal(t, "reg-1", gotReg.ID)
	assert.Len(t, gotReg.SelectedItems, 1)

	// 5. ListRegistrations
	regs, err := repo.ListRegistrations(ctx, "evt-1", "user-1")
	require.NoError(t, err)
	assert.Len(t, regs, 1)

	// 6. UpdateRegistration
	reg.Status = event.RegistrationStatusConfirmed
	require.NoError(t, repo.UpdateRegistration(ctx, reg))
	gotReg2, _ := repo.GetRegistration(ctx, "reg-1")
	assert.Equal(t, event.RegistrationStatusConfirmed, gotReg2.Status)
}

func TestEventRepository_ListFilters(t *testing.T) {
	db := newEventTestDB(t)
	repo := NewEventRepository(db)
	ctx := context.Background()

	_ = repo.Create(ctx, &event.Event{ID: "active", Status: event.EventStatusActive, CreatorID: "c1"})
	_ = repo.Create(ctx, &event.Event{ID: "draft", Status: event.EventStatusDraft, CreatorID: "c1"})

	t.Run("SysAdmin", func(t *testing.T) {
		res, _ := repo.List(ctx, 10, 0, "admin", true, false)
		assert.Len(t, res, 2)
	})

	t.Run("Manager_ManageOnly", func(t *testing.T) {
		res, _ := repo.List(ctx, 10, 0, "c1", false, true)
		assert.Len(t, res, 2)
	})

	t.Run("Public", func(t *testing.T) {
		res, _ := repo.List(ctx, 10, 0, "", false, false)
		assert.Len(t, res, 1)
		assert.Equal(t, "active", res[0].ID)
	})
}
