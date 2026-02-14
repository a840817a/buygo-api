package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/buygo/buygo-api/internal/adapter/repository/memory"
	"github.com/buygo/buygo-api/internal/domain/auth"
	"github.com/buygo/buygo-api/internal/domain/groupbuy"
	"github.com/buygo/buygo-api/internal/domain/user"
	"github.com/stretchr/testify/assert"
)

func TestProjectService_AccessControl(t *testing.T) {
	repo := memory.NewProjectRepository()
	svc := NewGroupBuyService(repo)

	// Contexts
	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	// managerCtx := auth.NewContext(context.Background(), "manager-1", int(user.UserRoleUser)) // Managers can be regular users role-wise, just assigned
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	anonCtx := context.Background()

	// 1. Create Project
	// Anon -> Fail
	_, err := svc.CreateProject(anonCtx, "Title", "Desc")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Anon should not create project, got %v", err)
	}

	// User -> Fail
	_, err = svc.CreateProject(userCtx, "Title", "Desc")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Regular User should not create project, got %v", err)
	}

	// Creator -> Success
	p, err := svc.CreateProject(creatorCtx, "My Project", "Desc")
	if err != nil {
		t.Fatalf("Creator should create project, got %v", err)
	}

	// 2. Update Project
	// Non-Manager User -> Fail (retained from original, as the provided snippet didn't explicitly remove it)
	_, err = svc.UpdateProject(userCtx, p.ID, "Updated Title", "Updated Desc", project.ProjectStatusActive, nil, "http://new.jpg", nil, nil, nil, 0, nil, "")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Non-Manager should not update project, got %v", err)
	}

	// Update (from provided snippet)
	now := time.Now()
	pUpdated, err := svc.UpdateProject(creatorCtx, p.ID, "New Title", "New Desc", project.ProjectStatusActive, nil, "http://new.com", &now, nil, nil, 0, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "New Title", pUpdated.Title)
	assert.Equal(t, project.ProjectStatusActive, pUpdated.Status)
	assert.Equal(t, "http://new.com", pUpdated.CoverImage)

	// Update with products (from provided snippet)
	prod := &project.Product{
		Name:          "Prod 1",
		PriceOriginal: 100,
		ExchangeRate:  0.25,
		MaxQuantity:   10,
	}
	pUpdated, err = svc.UpdateProject(creatorCtx, p.ID, "", "", project.ProjectStatusActive, []*project.Product{prod}, "", nil, nil, nil, 0, nil, "")
	if err != nil {
		t.Errorf("Creator/Manager should update project, got %v", err)
	}

	// 3. Get Project (Public)
	_, err = svc.GetProject(anonCtx, p.ID)
	if err != nil {
		t.Errorf("Anon should be able to get project, got %v", err)
	}

	// 4. Create Order
	// Anon -> Fail
	_, err = svc.CreateOrder(anonCtx, p.ID, nil, "Contact", "Addr", "", "")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Anon should not create order, got %v", err)
	}

	// User -> Success
	o, err := svc.CreateOrder(userCtx, p.ID, nil, "Contact", "Addr", "", "")
	if err != nil {
		t.Errorf("User should create order, got %v", err)
	}

	// 5. Cancel Order
	// Other User -> Fail
	otherUserCtx := auth.NewContext(context.Background(), "user-2", int(user.UserRoleUser))
	err = svc.CancelOrder(otherUserCtx, o.ID)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Other user should not cancel order, got %v", err)
	}

	// Owner -> Success
	err = svc.CancelOrder(userCtx, o.ID)
	if err != nil {
		t.Errorf("Owner should cancel order, got %v", err)
	}

	// 6. List Project Orders (Manager Only)
	// Anon -> Fail
	_, err = svc.ListProjectOrders(anonCtx, p.ID)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Anon should not list orders, got %v", err)
	}

	// User -> Fail
	_, err = svc.ListProjectOrders(userCtx, p.ID)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("User should not list orders, got %v", err)
	}

	// Creator/Manager -> Success
	orders, err := svc.ListProjectOrders(creatorCtx, p.ID)
	if err != nil {
		t.Errorf("Manager should list orders, got %v", err)
	}
	if len(orders) == 0 {
		// We expect at least the one created above?
		// Ah, repo is shared? yes `repo := memory.NewProjectRepository()`
		// But verification above cancelled it? No, CancelOrder doesn't delete, just updates status (mock impl)
		// Wait, Mock CancelOrder logic was commented out in Service?
		// "Need update repo logic ... return nil"
		// The CreateOrder logic saves it to repo. So it should be there.
		t.Errorf("Expected orders, got 0")
	}
}

func TestProjectService_ListPermissions(t *testing.T) {
	repo := memory.NewProjectRepository()
	svc := NewGroupBuyService(repo)

	// Contexts
	adminCtx := auth.NewContext(context.Background(), "admin", int(user.UserRoleSysAdmin))
	u1Ctx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleCreator))
	u2Ctx := auth.NewContext(context.Background(), "user-2", int(user.UserRoleCreator))
	publicCtx := context.Background()

	// Setup Data
	// P1: User1, Active
	p1 := &project.Project{ID: "p1", Title: "P1", Status: project.ProjectStatusActive, CreatorID: "user-1", ManagerIDs: []string{"user-1"}}
	repo.Create(adminCtx, p1)
	// P2: User1, Draft
	p2 := &project.Project{ID: "p2", Title: "P2", Status: project.ProjectStatusDraft, CreatorID: "user-1", ManagerIDs: []string{"user-1"}}
	repo.Create(adminCtx, p2)
	// P3: User2, Draft
	p3 := &project.Project{ID: "p3", Title: "P3", Status: project.ProjectStatusDraft, CreatorID: "user-2", ManagerIDs: []string{"user-2"}}
	repo.Create(adminCtx, p3)

	// 1. Public List: Should only see Active (P1)
	list, err := svc.ListProjects(publicCtx, 100, 0)
	if err != nil {
		t.Fatalf("Public list failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != "p1" {
		t.Errorf("Public list should return only P1, got %d items", len(list))
	}

	// 2. Manager List (User1): Should see own projects (P1, P2)
	list, err = svc.ListManagerProjects(u1Ctx, 100, 0)
	if err != nil {
		t.Fatalf("Manager list failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("User1 should see 2 projects, got %d", len(list))
	}
	// Verify IDs
	ids := make(map[string]bool)
	for _, p := range list {
		ids[p.ID] = true
	}
	if !ids["p1"] || !ids["p2"] {
		t.Errorf("User1 missing expected projects (P1, P2), got %v", list)
	}
	if ids["p3"] {
		t.Errorf("User1 sees User2's P3")
	}

	// 3. Manager List (User2): Should see own projects (P3)
	list, err = svc.ListManagerProjects(u2Ctx, 100, 0)
	if err != nil {
		t.Fatalf("Manager list (u2) failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != "p3" {
		t.Errorf("User2 should see P3, got %v", list)
	}

	// 4. Admin List: Should see all (P1, P2, P3)
	list, err = svc.ListManagerProjects(adminCtx, 100, 0)
	if err != nil {
		t.Fatalf("Admin list failed: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("Admin should see 3 projects, got %d", len(list))
	}

	// 5. Anon ListManagerProjects -> Fail
	_, err = svc.ListManagerProjects(publicCtx, 100, 0)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Anon ListManagerProjects should fail with PermissionDenied, got %v", err)
	}
}
