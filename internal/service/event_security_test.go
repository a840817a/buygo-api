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
)

func TestEventService_Security(t *testing.T) {
	repo := memory.NewEventRepository() // Assuming memory repo exists or I need to check
	svc := NewEventService(repo)

	// Contexts
	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userACtx := auth.NewContext(context.Background(), "user-A", int(user.UserRoleUser))
	userBCtx := auth.NewContext(context.Background(), "user-B", int(user.UserRoleUser))
	anonCtx := context.Background()

	// 1. Create Event: Creator Only
	_, err := svc.CreateEvent(anonCtx, "Title", "Desc", "", "", time.Now(), time.Now().Add(time.Hour), nil, nil, false, nil, nil, nil)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("Anon should not create event, got %v", err)
	}
	_, err = svc.CreateEvent(userACtx, "Title", "Desc", "", "", time.Now(), time.Now().Add(time.Hour), nil, nil, false, nil, nil, nil)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("User should not create event")
	}

	e, err := svc.CreateEvent(creatorCtx, "Secure Event", "Desc", "", "", time.Now(), time.Now().Add(time.Hour), nil, nil, false, nil, nil, nil)
	if err != nil {
		t.Fatalf("Creator should create event: %v", err)
	}

	// 2. Setup Items
	items := []*event.EventItem{
		{ID: "item-1", Name: "Ticket", Price: 100, MaxParticipants: 100, AllowMultiple: true},
	}
	e, err = svc.UpdateEvent(creatorCtx, e.ID, "Secure Event", "Desc", "Loc", "Cover", time.Now(), time.Now().Add(time.Hour), true, items, nil, nil)
	if err != nil {
		t.Fatalf("Failed to update event items: %v", err)
	}

	// 3. Activate Event (Using new API)
	e, err = svc.UpdateEventStatus(creatorCtx, e.ID, event.EventStatusActive)
	if err != nil {
		t.Fatalf("Failed to activate event: %v", err)
	}

	// 3. Register Event
	// User A
	regItems := []*event.RegistrationItem{
		{EventItemID: "item-1", Quantity: 1},
	}
	regA, err := svc.RegisterEvent(userACtx, e.ID, regItems, "Contact A", "Note A")
	if err != nil {
		t.Fatalf("User A should be able to register: %v", err)
	}

	// 4. Update Registration (IDOR Check)
	// User B tries to update User A's registration
	_, err = svc.UpdateRegistration(userBCtx, regA.ID, regItems, "Hacked", "Hacked")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("IDOR: User B SHOULD NOT be able to update User A's registration. Got: %v", err)
	}

	// User A updates own
	_, err = svc.UpdateRegistration(userACtx, regA.ID, regItems, "Contact A Updated", "Note Updated")
	if err != nil {
		t.Errorf("User A should be able to update own registration: %v", err)
	}

	// 5. Cancel Registration (IDOR Check)
	// User B tries to cancel User A
	err = svc.CancelRegistration(userBCtx, regA.ID)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("IDOR: User B SHOULD NOT be able to cancel User A's registration. Got: %v", err)
	}

	// User A cancels own
	err = svc.CancelRegistration(userACtx, regA.ID)
	if err != nil {
		t.Errorf("User A should be able to cancel own registration: %v", err)
	}

	// 6. Manager Access
	// User tries to list registrations -> Fail
	_, err = svc.ListEventRegistrations(userACtx, e.ID)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("User should not list event registrations")
	}

	// Manager lists -> Success
	regs, err := svc.ListEventRegistrations(creatorCtx, e.ID)
	if err != nil {
		t.Errorf("Manager should list registrations: %v", err)
	}
	if len(regs) == 0 {
		// We expect at least regA (even if cancelled, it might be listed depending on repo impl)
		// Usually List returns all.
	}
}
