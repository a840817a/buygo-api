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

func setupActiveEvent(t *testing.T) (*EventService, *event.Event, context.Context, context.Context) {
	t.Helper()
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))

	items := []*event.EventItem{
		{ID: "item-1", Name: "Ticket A", Price: 100, MaxParticipants: 50, AllowMultiple: true},
		{ID: "item-2", Name: "Ticket B", Price: 200, MaxParticipants: 30, AllowMultiple: false},
	}

	e, err := svc.CreateEvent(creatorCtx, "Test Event", "Desc", "", "",
		time.Now(), time.Now().Add(24*time.Hour), nil, nil, false, nil, items, nil)
	require.NoError(t, err)

	// Set registration deadline far in the future
	e.RegistrationDeadline = time.Now().Add(12 * time.Hour)
	e, err = svc.UpdateEvent(creatorCtx, e.ID, e.Title, e.Description, "", "", e.StartTime, e.EndTime, false, e.Items, nil, nil)
	require.NoError(t, err)

	// Activate
	e, err = svc.UpdateEventStatus(creatorCtx, e.ID, event.EventStatusActive)
	require.NoError(t, err)

	return svc, e, creatorCtx, userCtx
}

// --- RegisterEvent: Inactive Event ---

func TestRegisterEvent_InactiveEvent(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))

	e, _ := svc.CreateEvent(creatorCtx, "Draft Event", "Desc", "", "",
		time.Now(), time.Now().Add(time.Hour), nil, nil, false, nil, nil, nil)
	// Event stays in Draft status

	_, err := svc.RegisterEvent(userCtx, e.ID, nil, "C", "N")
	assert.Error(t, err, "Should not register for inactive event")
	assert.Contains(t, err.Error(), "not active")
}

// --- RegisterEvent: Deadline Passed ---

func TestRegisterEvent_DeadlinePassed(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))

	e, _ := svc.CreateEvent(creatorCtx, "Event", "Desc", "", "",
		time.Now(), time.Now().Add(time.Hour), nil, nil, false, nil, nil, nil)
	e, _ = svc.UpdateEventStatus(creatorCtx, e.ID, event.EventStatusActive)

	// Set deadline in the past
	e.RegistrationDeadline = time.Now().Add(-1 * time.Hour)
	repo.Update(context.Background(), e)

	_, err := svc.RegisterEvent(userCtx, e.ID, nil, "C", "N")
	assert.Error(t, err, "Should not register after deadline")
	assert.Contains(t, err.Error(), "deadline")
}

// --- RegisterEvent: Duplicate Registration ---

func TestRegisterEvent_Duplicate(t *testing.T) {
	svc, e, _, userCtx := setupActiveEvent(t)

	regItems := []*event.RegistrationItem{{EventItemID: "item-1", Quantity: 1}}

	// First registration → success
	_, err := svc.RegisterEvent(userCtx, e.ID, regItems, "C", "N")
	require.NoError(t, err)

	// Second registration → error (duplicate)
	_, err = svc.RegisterEvent(userCtx, e.ID, regItems, "C2", "N2")
	assert.Error(t, err, "Should not allow duplicate registration")
	assert.Contains(t, err.Error(), "already registered")
}

// --- RegisterEvent: Invalid Item ID ---

func TestRegisterEvent_InvalidItem(t *testing.T) {
	svc, e, _, userCtx := setupActiveEvent(t)

	regItems := []*event.RegistrationItem{{EventItemID: "non-existent", Quantity: 1}}
	_, err := svc.RegisterEvent(userCtx, e.ID, regItems, "C", "N")
	assert.Error(t, err, "Invalid item ID should error")
	assert.Contains(t, err.Error(), "invalid event item")
}

// --- RegisterEvent: Quantity Limit (!AllowMultiple) ---

func TestRegisterEvent_QuantityLimit(t *testing.T) {
	svc, e, _, userCtx := setupActiveEvent(t)

	// item-2 has AllowMultiple=false
	regItems := []*event.RegistrationItem{{EventItemID: "item-2", Quantity: 3}}
	_, err := svc.RegisterEvent(userCtx, e.ID, regItems, "C", "N")
	assert.Error(t, err, "Exceeded quantity limit should error")
	assert.Contains(t, err.Error(), "quantity limit")

	// Quantity 1 should work
	regItems = []*event.RegistrationItem{{EventItemID: "item-2", Quantity: 1}}
	_, err = svc.RegisterEvent(userCtx, e.ID, regItems, "C", "N")
	assert.NoError(t, err)
}

// --- RegisterEvent: Anon Denied ---

func TestRegisterEvent_AnonDenied(t *testing.T) {
	svc, e, _, _ := setupActiveEvent(t)
	anonCtx := context.Background()

	_, err := svc.RegisterEvent(anonCtx, e.ID, nil, "C", "N")
	assert.True(t, errors.Is(err, ErrUnauthorized))
}

// --- GetMyRegistrations ---

func TestGetMyRegistrations(t *testing.T) {
	svc, e, _, userCtx := setupActiveEvent(t)
	anonCtx := context.Background()
	userBCtx := auth.NewContext(context.Background(), "user-b", int(user.UserRoleUser))

	// Anon → denied
	_, err := svc.GetMyRegistrations(anonCtx)
	assert.True(t, errors.Is(err, ErrUnauthorized))

	// No registrations → empty
	regs, err := svc.GetMyRegistrations(userCtx)
	assert.NoError(t, err)
	assert.Empty(t, regs)

	// Register → returns own
	regItems := []*event.RegistrationItem{{EventItemID: "item-1", Quantity: 1}}
	svc.RegisterEvent(userCtx, e.ID, regItems, "C", "N")

	regs, err = svc.GetMyRegistrations(userCtx)
	assert.NoError(t, err)
	assert.Len(t, regs, 1)

	// Other user sees empty
	regsB, err := svc.GetMyRegistrations(userBCtx)
	assert.NoError(t, err)
	assert.Empty(t, regsB)
}

// --- UpdateRegistration: Items Changed → Status Reset ---

func TestUpdateRegistration_ItemsChangedStatusReset(t *testing.T) {
	svc, e, creatorCtx, userCtx := setupActiveEvent(t)

	regItems := []*event.RegistrationItem{{EventItemID: "item-1", Quantity: 1}}
	reg, _ := svc.RegisterEvent(userCtx, e.ID, regItems, "C", "N")

	// Manager confirms the registration
	reg, _ = svc.UpdateRegistrationStatus(creatorCtx, reg.ID, event.RegistrationStatusConfirmed, 0)
	assert.Equal(t, event.RegistrationStatusConfirmed, reg.Status)

	// User changes items → status should reset to Pending
	newItems := []*event.RegistrationItem{{EventItemID: "item-1", Quantity: 2}}
	updated, err := svc.UpdateRegistration(userCtx, reg.ID, newItems, "C2", "N2")
	require.NoError(t, err)
	assert.Equal(t, event.RegistrationStatusPending, updated.Status, "Status should reset when items change")
}
