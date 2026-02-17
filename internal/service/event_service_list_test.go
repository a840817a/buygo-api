package service

import (
	"context"
	"errors"
	"testing"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/memory"
	"github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

func TestEventService_ListPermissions(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := NewEventService(repo)

	// Contexts
	adminCtx := auth.NewContext(context.Background(), "admin", int(user.UserRoleSysAdmin))
	u1Ctx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleCreator))
	u2Ctx := auth.NewContext(context.Background(), "user-2", int(user.UserRoleCreator))
	publicCtx := context.Background()

	// Setup Data
	// E1: User1, Active
	e1 := &event.Event{ID: "e1", Title: "E1", Status: event.EventStatusActive, CreatorID: "user-1", ManagerIDs: []string{"user-1"}}
	repo.Create(adminCtx, e1)
	// E2: User1, Draft
	e2 := &event.Event{ID: "e2", Title: "E2", Status: event.EventStatusDraft, CreatorID: "user-1", ManagerIDs: []string{"user-1"}}
	repo.Create(adminCtx, e2)
	// E3: User2, Draft
	e3 := &event.Event{ID: "e3", Title: "E3", Status: event.EventStatusDraft, CreatorID: "user-2", ManagerIDs: []string{"user-2"}}
	repo.Create(adminCtx, e3)

	// 1. Public List: Should only see Active (E1)
	list, err := svc.ListEvents(publicCtx, 100, 0)
	if err != nil {
		t.Fatalf("Public list failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != "e1" {
		t.Errorf("Public list should return only E1, got %d items", len(list))
	}

	// 2. Manager List (User1): Should see own events (E1, E2)
	list, err = svc.ListManagerEvents(u1Ctx, 100, 0)
	if err != nil {
		t.Fatalf("Manager list failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("User1 should see 2 events, got %d", len(list))
	}
	ids := make(map[string]bool)
	for _, e := range list {
		ids[e.ID] = true
	}
	if !ids["e1"] || !ids["e2"] {
		t.Errorf("User1 missing expected events (E1, E2), got %v", list)
	}

	// 3. Manager List (User2): Should see own events (E3)
	list, err = svc.ListManagerEvents(u2Ctx, 100, 0)
	if err != nil {
		t.Fatalf("Manager list (u2) failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != "e3" {
		t.Errorf("User2 should see E3, got %v", list)
	}

	// 4. Admin List: Should see all (E1, E2, E3)
	list, err = svc.ListManagerEvents(adminCtx, 100, 0)
	if err != nil {
		t.Fatalf("Admin list failed: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("Admin should see 3 events, got %d", len(list))
	}

	// 5. Anon ListManagerEvents -> Fail
	_, err = svc.ListManagerEvents(publicCtx, 100, 0)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("Anon ListManagerEvents should fail with ErrUnauthorized, got %v", err)
	}
}
