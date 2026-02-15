package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/buygo/buygo-api/internal/adapter/repository/memory"
	"github.com/buygo/buygo-api/internal/domain/auth"
	"github.com/buygo/buygo-api/internal/domain/event"
	"github.com/buygo/buygo-api/internal/domain/user"
)

func TestEventService_Logic_UpdateRegistration(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	// Contexts
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	otherCtx := auth.NewContext(context.Background(), "user-2", int(user.UserRoleUser))
	// adminCtx := auth.NewContext(context.Background(), "admin", int(user.UserRoleSysAdmin))

	// 1. Setup Event (Active, Future Deadline)
	e := &event.Event{
		ID:                   "event-1",
		Status:               event.EventStatusActive,
		RegistrationDeadline: time.Now().Add(1 * time.Hour),
		Items: []*event.EventItem{
			{ID: "item-1", Price: 100},
		},
		AllowException: false,
	}
	repo.Create(context.Background(), e)

	// 2. Register
	// Pass empty notes as per updated signature
	reg, err := svc.RegisterEvent(userCtx, e.ID, []*event.RegistrationItem{{EventItemID: "item-1", Quantity: 1}}, "Contact", "Initial Note")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 3. Update (Normal - Before Deadline)
	// Update Notes
	updReg, err := svc.UpdateRegistration(userCtx, reg.ID, reg.SelectedItems, "Contact Updated", "Updated Note")
	if err != nil {
		t.Fatalf("Update before deadline failed: %v", err)
	}
	if updReg.Notes != "Updated Note" {
		t.Errorf("Expected note updated, got %s", updReg.Notes)
	}

	// 4. Update (Different User - Fail)
	_, err = svc.UpdateRegistration(otherCtx, reg.ID, reg.SelectedItems, "", "")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Other user should not update, got %v", err)
	}

	// 5. Update (Deadline Passed, AllowException = False)
	e.RegistrationDeadline = time.Now().Add(-1 * time.Hour)
	repo.Update(context.Background(), e)

	_, err = svc.UpdateRegistration(userCtx, reg.ID, reg.SelectedItems, "Late", "Late")
	if err == nil {
		t.Error("Update should fail after deadline (AllowException=false)")
	}

	// 6. Update (Deadline Passed, AllowException = True)
	e.AllowException = true
	repo.Update(context.Background(), e)

	updReg, err = svc.UpdateRegistration(userCtx, reg.ID, reg.SelectedItems, "Contact Late", "Note Late")
	if err != nil {
		t.Errorf("Update should succeed after deadline if AllowException=true, got %v", err)
	}
	if updReg.Notes != "Note Late" {
		t.Errorf("Expected note updated (exception), got %s", updReg.Notes)
	}
}

func TestEventService_Logic_UpdateStatus(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	creatorID := "creator-1"
	managerID := "manager-1"
	userID := "user-1"

	// Contexts
	creatorCtx := auth.NewContext(context.Background(), creatorID, int(user.UserRoleCreator))
	managerCtx := auth.NewContext(context.Background(), managerID, int(user.UserRoleUser)) // even pure user role but in managers list should work? Interface implies role check or list check.
	// Service logic:
	// if role == SysAdmin -> OK
	// else check if creator OR in managers list.
	// We need to ensure logic handles role check properly.

	userCtx := auth.NewContext(context.Background(), userID, int(user.UserRoleUser))
	adminCtx := auth.NewContext(context.Background(), "admin", int(user.UserRoleSysAdmin))

	// Setup Event
	e := &event.Event{
		ID:         "event-1",
		CreatorID:  creatorID,
		ManagerIDs: []string{managerID},
		Status:     event.EventStatusActive,
	}
	repo.Create(context.Background(), e)

	// Register
	reg, _ := svc.RegisterEvent(userCtx, e.ID, nil, "C", "N")

	// 1. User Update Status -> Fail
	_, err := svc.UpdateRegistrationStatus(userCtx, reg.ID, event.RegistrationStatusConfirmed, 0)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("User should not update status, got %v", err)
	}

	// 2. Creator Update -> Success
	upd, err := svc.UpdateRegistrationStatus(creatorCtx, reg.ID, event.RegistrationStatusConfirmed, event.PaymentStatusSubmitted)
	if err != nil {
		t.Errorf("Creator should update status, got %v", err)
	}
	if upd.Status != event.RegistrationStatusConfirmed || upd.PaymentStatus != event.PaymentStatusSubmitted {
		t.Errorf("Status mismatch")
	}

	// 3. Manager Update -> Success
	// Reset
	reg.PaymentStatus = event.PaymentStatusUnpaid
	repo.UpdateRegistration(context.Background(), reg)

	upd, err = svc.UpdateRegistrationStatus(managerCtx, reg.ID, 0, event.PaymentStatusPaid)
	if err != nil {
		t.Errorf("Manager should update status, got %v", err)
	}
	// Status should remain confirmed (from previous step) if passed 0/Unspecified
	// Wait, Unspecified is 0. Service check: "if status != 0 { update }"
	if upd.Status != event.RegistrationStatusConfirmed {
		t.Errorf("Status should persist if unspecified passed")
	}
	if upd.PaymentStatus != event.PaymentStatusPaid {
		t.Errorf("Payment status should update")
	}

	// 4. Admin Update -> Success
	upd, err = svc.UpdateRegistrationStatus(adminCtx, reg.ID, event.RegistrationStatusCancelled, event.PaymentStatusRefunded)
	if err != nil {
		t.Errorf("Admin should update status, got %v", err)
	}
	if upd.Status != event.RegistrationStatusCancelled {
		t.Errorf("Admin update failed")
	}
}
