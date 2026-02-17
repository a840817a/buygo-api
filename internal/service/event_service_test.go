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

func TestEventService_AccessControl(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	// Contexts
	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	anonCtx := context.Background()

	// 1. Create Event
	// Anon -> Fail
	_, err := svc.CreateEvent(anonCtx, "Title", "Desc", "", "", time.Now(), time.Now(), nil, nil, false, nil, nil, nil)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Anon should not create event, got %v", err)
	}

	// Creator -> Success
	e, err := svc.CreateEvent(creatorCtx, "My Event", "Desc", "", "", time.Now(), time.Now(), nil, nil, false, nil, nil, nil)
	if err != nil {
		t.Fatalf("Creator should create event, got %v", err)
	}

	// Activate event for registration (Hack: Update direct or expose method)
	// For test, let's assume Create makes it Draft, need to simulate Active for registration
	// Since we don't have Activate RPC explicitly yet (Update), we can use Update directly on Repo or assume success if logic permits
	// Actually, CreateEvent sets Draft. Register checks Active. We need to Update it.
	// We didn't impl UpdateEvent in Service yet! Let's just manually update repo for test
	e.Status = event.EventStatusActive
	repo.Update(context.Background(), e)

	// 2. List Events (Public)
	_, err = svc.ListEvents(anonCtx, 10, 0)
	if err != nil {
		t.Errorf("Anon should list events, got %v", err)
	}

	// 3. Register Event
	// Anon -> Fail
	_, err = svc.RegisterEvent(anonCtx, e.ID, nil, "Contact", "")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Anon should not register, got %v", err)
	}

	// User -> Success
	r, err := svc.RegisterEvent(userCtx, e.ID, nil, "Contact", "")
	if err != nil {
		t.Errorf("User should register, got %v", err)
	}

	// 4. Cancel Registration
	// Other User -> Fail
	otherUserCtx := auth.NewContext(context.Background(), "user-2", int(user.UserRoleUser))
	err = svc.CancelRegistration(otherUserCtx, r.ID)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Other user should not cancel reg, got %v", err)
	}

	// Owner -> Success
	err = svc.CancelRegistration(userCtx, r.ID)
	if err != nil {
		t.Errorf("Owner should cancel reg, got %v", err)
	}

	// 5. List Event Registrations
	// Normal User -> Fail
	_, err = svc.ListEventRegistrations(userCtx, e.ID)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("User should not list event registrations, got %v", err)
	}

	// Creator -> Success
	_, err = svc.ListEventRegistrations(creatorCtx, e.ID)
	if err != nil {
		t.Errorf("Creator should list event registrations, got %v", err)
	}
}
