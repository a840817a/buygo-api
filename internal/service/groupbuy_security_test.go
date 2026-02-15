package service

import (
	"context"
	"errors"
	"testing"

	"github.com/buygo/buygo-api/internal/adapter/repository/memory"
	"github.com/buygo/buygo-api/internal/domain/auth"
	"github.com/buygo/buygo-api/internal/domain/groupbuy"
	"github.com/buygo/buygo-api/internal/domain/user"
)

func TestGroupBuyService_IDOR_Prevention(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	// Contexts
	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userACtx := auth.NewContext(context.Background(), "user-A", int(user.UserRoleUser))
	userBCtx := auth.NewContext(context.Background(), "user-B", int(user.UserRoleUser))

	// 1. Setup: Create GroupBuy and Order for User A
	gb, err := svc.CreateGroupBuy(creatorCtx, "IDOR Test GroupBuy", "Desc")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Activate GroupBuy
	_, err = svc.UpdateGroupBuy(creatorCtx, gb.ID, "", "", groupbuy.GroupBuyStatusActive, nil, "", nil, nil, nil, 0, nil, "")
	if err != nil {
		t.Fatalf("Failed to activate project: %v", err)
	}

	orderA, err := svc.CreateOrder(userACtx, gb.ID, nil, "Contact A", "Addr A", "", "")
	if err != nil {
		t.Fatalf("User A failed to create order: %v", err)
	}

	// 2. IDOR Attack: User B tries to update User A's order payment info
	_, err = svc.UpdatePaymentInfo(userBCtx, orderA.ID, "Hack", "00000", "", "", nil, 0)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("IDOR: User B SHOULD NOT be able to update User A's payment info. Got: %v", err)
	}

	// 3. IDOR Attack: User B tries to update User A's order items
	_, err = svc.UpdateOrder(userBCtx, orderA.ID, nil, "")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("IDOR: User B SHOULD NOT be able to update User A's order items. Got: %v", err)
	}

	// 4. Verfiy User A CAN update their own order
	// UpdatePaymentInfo should succeed for owner
	_, err = svc.UpdatePaymentInfo(userACtx, orderA.ID, "Legit", "12345", "New Contact", "", nil, 0)
	if err != nil {
		t.Errorf("Owner (User A) SHOULD be able to update their own order. Got: %v", err)
	}

	// 5. Verify Manager CAN update User A's order
	_, err = svc.UpdatePaymentInfo(creatorCtx, orderA.ID, "Manager Edit", "99999", "", "", nil, 0)
	if err != nil {
		t.Errorf("Manager SHOULD be able to update User A's order. Got: %v", err)
	}

	// 6. Verify GetMyOrders Isolation
	// Create an order for User B as well
	_, err = svc.CreateOrder(userBCtx, gb.ID, nil, "Contact B", "Addr B", "", "")
	if err != nil {
		t.Fatalf("User B failed to create order: %v", err)
	}

	// User A should only see 1 order (their own)
	ordersA, err := svc.GetMyOrders(userACtx)
	if err != nil {
		t.Fatalf("GetMyOrders failed for User A: %v", err)
	}
	if len(ordersA) != 1 {
		t.Errorf("User A should see exactly 1 order, got %d", len(ordersA))
	}
	if ordersA[0].ID != orderA.ID {
		t.Errorf("User A saw wrong order ID: %s", ordersA[0].ID)
	}

	// User B should only see 1 order (their own)
	ordersB, err := svc.GetMyOrders(userBCtx)
	if err != nil {
		t.Fatalf("GetMyOrders failed for User B: %v", err)
	}
	if len(ordersB) != 1 {
		t.Errorf("User B should see exactly 1 order, got %d", len(ordersB))
	}

	// 7. Verify Read by ID is not possible (Secure by Design)
	// There is no GetOrder(ctx, id) RPC exposed to users in the ProjectService interface.
	// Users can only access orders via GetMyOrders (list own) or GetMyGroupBuyOrder (by group buy ID, returns own).
	// Thus, it is impossible for User B to request User A's order by ID.
}
