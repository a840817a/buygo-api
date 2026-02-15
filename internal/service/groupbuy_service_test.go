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

func TestGroupBuyService_AccessControl(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	// Contexts
	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	// managerCtx := auth.NewContext(context.Background(), "manager-1", int(user.UserRoleUser)) // Managers can be regular users role-wise, just assigned
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	anonCtx := context.Background()

	// 1. Create GroupBuy
	// Anon -> Fail
	_, err := svc.CreateGroupBuy(anonCtx, "Title", "Desc")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Anon should not create project, got %v", err)
	}

	// User -> Fail
	_, err = svc.CreateGroupBuy(userCtx, "Title", "Desc")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Regular User should not create project, got %v", err)
	}

	// Creator -> Success
	gb, err := svc.CreateGroupBuy(creatorCtx, "My GroupBuy", "Desc")
	if err != nil {
		t.Fatalf("Creator should create project, got %v", err)
	}

	// 2. Update GroupBuy
	// Non-Manager User -> Fail (retained from original, as the provided snippet didn't explicitly remove it)
	_, err = svc.UpdateGroupBuy(userCtx, gb.ID, "Updated Title", "Updated Desc", groupbuy.GroupBuyStatusActive, nil, "http://new.jpg", nil, nil, nil, 0, nil, "")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Non-Manager should not update project, got %v", err)
	}

	// Update (from provided snippet)
	now := time.Now()
	gbUpdated, err := svc.UpdateGroupBuy(creatorCtx, gb.ID, "New Title", "New Desc", groupbuy.GroupBuyStatusActive, nil, "http://new.com", &now, nil, nil, 0, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, "New Title", gbUpdated.Title)
	assert.Equal(t, groupbuy.GroupBuyStatusActive, gbUpdated.Status)
	assert.Equal(t, "http://new.com", gbUpdated.CoverImage)

	// Update with products (from provided snippet)
	prod := &groupbuy.Product{
		Name:          "Prod 1",
		PriceOriginal: 100,
		ExchangeRate:  0.25,
		MaxQuantity:   10,
	}
	gbUpdated, err = svc.UpdateGroupBuy(creatorCtx, gb.ID, "", "", groupbuy.GroupBuyStatusActive, []*groupbuy.Product{prod}, "", nil, nil, nil, 0, nil, "")
	if err != nil {
		t.Errorf("Creator/Manager should update project, got %v", err)
	}

	// 3. Get GroupBuy (Public)
	_, err = svc.GetGroupBuy(anonCtx, gb.ID)
	if err != nil {
		t.Errorf("Anon should be able to get project, got %v", err)
	}

	// 4. Create Order
	// Anon -> Fail
	_, err = svc.CreateOrder(anonCtx, gb.ID, nil, "Contact", "Addr", "", "")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Anon should not create order, got %v", err)
	}

	// User -> Success
	o, err := svc.CreateOrder(userCtx, gb.ID, nil, "Contact", "Addr", "", "")
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

	// 6. List GroupBuy Orders (Manager Only)
	// Anon -> Fail
	_, err = svc.ListGroupBuyOrders(anonCtx, gb.ID)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Anon should not list orders, got %v", err)
	}

	// User -> Fail
	_, err = svc.ListGroupBuyOrders(userCtx, gb.ID)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("User should not list orders, got %v", err)
	}

	// Creator/Manager -> Success
	orders, err := svc.ListGroupBuyOrders(creatorCtx, gb.ID)
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

func TestGroupBuyService_ListPermissions(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	// Contexts
	adminCtx := auth.NewContext(context.Background(), "admin", int(user.UserRoleSysAdmin))
	u1Ctx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleCreator))
	u2Ctx := auth.NewContext(context.Background(), "user-2", int(user.UserRoleCreator))
	publicCtx := context.Background()

	// Setup Data
	// P1: User1, Active
	gb1 := &groupbuy.GroupBuy{ID: "gb1", Title: "GB1", Status: groupbuy.GroupBuyStatusActive, CreatorID: "user-1", ManagerIDs: []string{"user-1"}}
	repo.Create(adminCtx, gb1)
	// P2: User1, Draft
	gb2 := &groupbuy.GroupBuy{ID: "gb2", Title: "GB2", Status: groupbuy.GroupBuyStatusDraft, CreatorID: "user-1", ManagerIDs: []string{"user-1"}}
	repo.Create(adminCtx, gb2)
	// P3: User2, Draft
	gb3 := &groupbuy.GroupBuy{ID: "gb3", Title: "GB3", Status: groupbuy.GroupBuyStatusDraft, CreatorID: "user-2", ManagerIDs: []string{"user-2"}}
	repo.Create(adminCtx, gb3)

	// 1. Public List: Should only see Active (GB1)
	list, err := svc.ListGroupBuys(publicCtx, 100, 0)
	if err != nil {
		t.Fatalf("Public list failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != "gb1" {
		t.Errorf("Public list should return only GB1, got %d items", len(list))
	}

	// 2. Manager List (User1): Should see own projects (GB1, GB2)
	list, err = svc.ListManagerGroupBuys(u1Ctx, 100, 0)
	if err != nil {
		t.Fatalf("Manager list failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("User1 should see 2 group buys, got %d", len(list))
	}
	// Verify IDs
	ids := make(map[string]bool)
	for _, gb := range list {
		ids[gb.ID] = true
	}
	if !ids["gb1"] || !ids["gb2"] {
		t.Errorf("User1 missing expected group buys (GB1, GB2), got %v", list)
	}
	if ids["gb3"] {
		t.Errorf("User1 sees User2's GB3")
	}

	// 3. Manager List (User2): Should see own group buys (GB3)
	list, err = svc.ListManagerGroupBuys(u2Ctx, 100, 0)
	if err != nil {
		t.Fatalf("Manager list (u2) failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != "gb3" {
		t.Errorf("User2 should see P3, got %v", list)
	}

	// 4. Admin List: Should see all (GB1, GB2, GB3)
	list, err = svc.ListManagerGroupBuys(adminCtx, 100, 0)
	if err != nil {
		t.Fatalf("Admin list failed: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("Admin should see 3 group buys, got %d", len(list))
	}

	// 5. Anon ListManagerGroupBuys -> Fail
	_, err = svc.ListManagerGroupBuys(publicCtx, 100, 0)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("Anon ListManagerGroupBuys should fail with PermissionDenied, got %v", err)
	}
}
