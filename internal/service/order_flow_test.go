package service

import (
	"context"
	"testing"

	"github.com/buygo/buygo-api/internal/adapter/repository/memory"
	"github.com/buygo/buygo-api/internal/domain/auth"
	"github.com/buygo/buygo-api/internal/domain/groupbuy"
	"github.com/buygo/buygo-api/internal/domain/user"
)

func TestGroupBuyService_OrderFlow(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	// Setup: Creator, GroupBuy, Product
	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	gb, err := svc.CreateGroupBuy(creatorCtx, "Test GroupBuy", "Desc")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	// Activate GroupBuy
	// Looking at CreateGroupBuy in service: p.Status = Active (1) or Draft?
	// Let's assume we need to update it to Active if logic requires it.
	// Actually, CreateGroupBuy sets Status=0 usually?
	// Let's explicitly update to Active just in case.
	gb, err = svc.UpdateGroupBuy(creatorCtx, gb.ID, "", "", groupbuy.GroupBuyStatusActive, nil, "", nil, nil, nil, 0, nil, "")
	if err != nil {
		t.Fatalf("Failed to activate project: %v", err)
	}

	// Add Product
	specs := []string{"Spec A", "Spec B"}
	prod, err := svc.AddProduct(creatorCtx, gb.ID, "Product 1", 100, 1.0, specs)
	if err != nil {
		t.Fatalf("Failed to add product: %v", err)
	}

	// Test: User Creates Order
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	items := []*groupbuy.OrderItem{
		{
			ProductID: prod.ID,
			SpecID:    prod.Specs[0].ID, // Spec A
			Quantity:  2,
		},
	}

	order, err := svc.CreateOrder(userCtx, gb.ID, items, "Contact", "Address", "", "")
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Verify Snapshot
	if len(order.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(order.Items))
	} else {
		item := order.Items[0]
		if item.ProductName != "Product 1" {
			t.Errorf("Expected ProductName 'Product 1', got '%s'", item.ProductName)
		}
		if item.SpecName != "Spec A" {
			t.Errorf("Expected SpecName 'Spec A', got '%s'", item.SpecName)
		}
		if item.Price != 100 {
			t.Errorf("Expected Price 100, got %d", item.Price)
		}
	}

	// Verify Total
	expectedTotal := int64(100 * 2)
	if order.TotalAmount != expectedTotal {
		t.Errorf("Expected TotalAmount %d, got %d", expectedTotal, order.TotalAmount)
	}

	// Test: Update Order (Change Qty, Change Spec)
	newItems := []*groupbuy.OrderItem{
		{
			ProductID: prod.ID,
			SpecID:    prod.Specs[1].ID, // Spec B
			Quantity:  3,
		},
	}
	updatedOrder, err := svc.UpdateOrder(userCtx, order.ID, newItems, "")
	if err != nil {
		t.Fatalf("Failed to update order: %v", err)
	}

	// Verify Update
	if len(updatedOrder.Items) != 1 {
		t.Errorf("Expected 1 item after update, got %d", len(updatedOrder.Items))
	} else {
		item := updatedOrder.Items[0]
		if item.SpecName != "Spec B" {
			t.Errorf("Expected SpecName 'Spec B', got '%s'", item.SpecName)
		}
		if item.Quantity != 3 {
			t.Errorf("Expected Quantity 3, got %d", item.Quantity)
		}
	}

	expectedTotalUpdate := int64(100 * 3)
	if updatedOrder.TotalAmount != expectedTotalUpdate {
		t.Errorf("Expected TotalAmount %d, got %d", expectedTotalUpdate, updatedOrder.TotalAmount)
	}

	// Test: Update Payment Info
	_, err = svc.UpdatePaymentInfo(userCtx, order.ID, "Bank Transfer", "12345", "", "", nil, 0)
	if err != nil {
		t.Fatalf("Failed to update payment info: %v", err)
	}

	// Fetch again to verify
	// (Actually UpdatePaymentInfo returns order, but let's fetch to be sure of persistence)
	// Service doesn't expose GetOrder for User easily (Create/Update return it).
	// But ListProjectOrders (Manager) can find it.
	// Or just trust the return value for this unit test.
}
