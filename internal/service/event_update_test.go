package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/memory"
	"github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- UpdateEvent ---

func TestUpdateEvent_FieldsUpdate(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	e, _ := svc.CreateEvent(creatorCtx, "Original", "Desc", "", "",
		time.Now(), time.Now().Add(time.Hour), nil, nil, false, nil, nil, nil)

	newItems := []*event.EventItem{
		{Name: "VIP Ticket", Price: 500, AllowMultiple: true},
	}

	updated, err := svc.UpdateEvent(creatorCtx, e.ID, "New Title", "New Desc", "Tokyo", "http://cover.jpg",
		time.Now(), time.Now().Add(2*time.Hour), true, newItems, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "New Title", updated.Title)
	assert.Equal(t, "New Desc", updated.Description)
	assert.Equal(t, "Tokyo", updated.Location)
	assert.Equal(t, "http://cover.jpg", updated.CoverImage)
	assert.True(t, updated.AllowException)
	assert.Len(t, updated.Items, 1)
	assert.NotEmpty(t, updated.Items[0].ID, "Item should get auto-generated ID")
	assert.Equal(t, e.ID, updated.Items[0].EventID)
}

func TestUpdateEvent_NonManagerDenied(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	anonCtx := context.Background()

	e, _ := svc.CreateEvent(creatorCtx, "Event", "Desc", "", "",
		time.Now(), time.Now().Add(time.Hour), nil, nil, false, nil, nil, nil)

	// User → denied
	_, err := svc.UpdateEvent(userCtx, e.ID, "Hack", "", "", "",
		time.Now(), time.Now().Add(time.Hour), false, nil, nil, nil)
	assert.True(t, errors.Is(err, ErrPermissionDenied))

	// Anon → denied
	_, err = svc.UpdateEvent(anonCtx, e.ID, "Hack", "", "", "",
		time.Now(), time.Now().Add(time.Hour), false, nil, nil, nil)
	assert.True(t, errors.Is(err, ErrUnauthorized))
}

func TestUpdateEvent_OnlyCreatorUpdatesManagers(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))

	e, _ := svc.CreateEvent(creatorCtx, "Event", "Desc", "", "",
		time.Now(), time.Now().Add(time.Hour), nil, nil, false, nil, nil, nil)

	// Add a manager
	e, err := svc.UpdateEvent(creatorCtx, e.ID, e.Title, e.Description, "", "",
		e.StartTime, e.EndTime, false, nil, []string{"creator-1", "manager-new"}, nil)
	require.NoError(t, err)
	assert.Contains(t, e.ManagerIDs, "manager-new")

	// Non-creator manager tries to update ManagerIDs → silently ignored
	managerCtx := auth.NewContext(context.Background(), "manager-new", int(user.UserRoleCreator))
	e, err = svc.UpdateEvent(managerCtx, e.ID, e.Title, e.Description, "", "",
		e.StartTime, e.EndTime, false, nil, []string{"manager-new"}, nil)
	require.NoError(t, err)
	// ManagerIDs should NOT have changed (only creator can update)
	assert.Contains(t, e.ManagerIDs, "creator-1", "Original creator should still be in managers")
}

// --- UpdateEventStatus ---

func TestUpdateEventStatus_Transitions(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))

	e, _ := svc.CreateEvent(creatorCtx, "Event", "Desc", "", "",
		time.Now(), time.Now().Add(time.Hour), nil, nil, false, nil, nil, nil)
	assert.Equal(t, event.EventStatusDraft, e.Status)

	// Draft → Active
	e, err := svc.UpdateEventStatus(creatorCtx, e.ID, event.EventStatusActive)
	require.NoError(t, err)
	assert.Equal(t, event.EventStatusActive, e.Status)

	// Active → Ended
	e, err = svc.UpdateEventStatus(creatorCtx, e.ID, event.EventStatusEnded)
	require.NoError(t, err)
	assert.Equal(t, event.EventStatusEnded, e.Status)
}

func TestUpdateEventStatus_NonManagerDenied(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	anonCtx := context.Background()

	e, _ := svc.CreateEvent(creatorCtx, "Event", "Desc", "", "",
		time.Now(), time.Now().Add(time.Hour), nil, nil, false, nil, nil, nil)

	// User → denied
	_, err := svc.UpdateEventStatus(userCtx, e.ID, event.EventStatusActive)
	assert.True(t, errors.Is(err, ErrPermissionDenied))

	// Anon → denied
	_, err = svc.UpdateEventStatus(anonCtx, e.ID, event.EventStatusActive)
	assert.True(t, errors.Is(err, ErrUnauthorized))
}

func TestUpdateEventStatus_SysAdminAllowed(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	adminCtx := auth.NewContext(context.Background(), "admin-1", int(user.UserRoleSysAdmin))

	e, _ := svc.CreateEvent(creatorCtx, "Event", "Desc", "", "",
		time.Now(), time.Now().Add(time.Hour), nil, nil, false, nil, nil, nil)

	// Admin can change status even though not creator
	e, err := svc.UpdateEventStatus(adminCtx, e.ID, event.EventStatusActive)
	require.NoError(t, err)
	assert.Equal(t, event.EventStatusActive, e.Status)
}
