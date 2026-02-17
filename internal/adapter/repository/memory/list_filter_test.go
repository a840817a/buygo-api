package memory

import (
	"context"
	"testing"
	"time"

	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

func TestUserRepository_List_Pagination(t *testing.T) {
	repo := NewUserRepository()
	now := time.Now()
	_ = repo.Create(context.Background(), &user.User{ID: "u-1", CreatedAt: now, UpdatedAt: now})
	_ = repo.Create(context.Background(), &user.User{ID: "u-2", CreatedAt: now, UpdatedAt: now})
	_ = repo.Create(context.Background(), &user.User{ID: "u-3", CreatedAt: now, UpdatedAt: now})

	first, err := repo.List(context.Background(), 2, 0)
	if err != nil {
		t.Fatalf("list first page error: %v", err)
	}
	if len(first) != 2 {
		t.Fatalf("first page len = %d, want 2", len(first))
	}

	overflow, err := repo.List(context.Background(), 2, 10)
	if err != nil {
		t.Fatalf("list overflow page error: %v", err)
	}
	if len(overflow) != 0 {
		t.Fatalf("overflow len = %d, want 0", len(overflow))
	}
}

func TestGroupBuyRepository_List_FilterAndPaging(t *testing.T) {
	repo := NewGroupBuyRepository()
	now := time.Now()

	_ = repo.Create(context.Background(), &groupbuy.GroupBuy{
		ID:         "gb-draft-managed",
		Title:      "Draft Managed",
		Status:     groupbuy.GroupBuyStatusDraft,
		CreatorID:  "creator-1",
		ManagerIDs: []string{"manager-1"},
		CreatedAt:  now.Add(-1 * time.Hour),
	})
	_ = repo.Create(context.Background(), &groupbuy.GroupBuy{
		ID:        "gb-active-public",
		Title:     "Active Public",
		Status:    groupbuy.GroupBuyStatusActive,
		CreatedAt: now.Add(-2 * time.Hour),
	})
	_ = repo.Create(context.Background(), &groupbuy.GroupBuy{
		ID:        "gb-ended-public",
		Title:     "Ended Public",
		Status:    groupbuy.GroupBuyStatusEnded,
		CreatedAt: now.Add(-3 * time.Hour),
	})

	publicList, err := repo.List(context.Background(), 10, 0, "", false, false)
	if err != nil {
		t.Fatalf("public list error: %v", err)
	}
	if len(publicList) != 2 {
		t.Fatalf("public list len = %d, want 2", len(publicList))
	}
	if publicList[0].ID != "gb-active-public" {
		t.Fatalf("public list order[0] = %q, want %q", publicList[0].ID, "gb-active-public")
	}

	managerList, err := repo.List(context.Background(), 10, 0, "manager-1", false, true)
	if err != nil {
		t.Fatalf("manager list error: %v", err)
	}
	if len(managerList) != 1 || managerList[0].ID != "gb-draft-managed" {
		t.Fatalf("manager list mismatch: %+v", managerList)
	}

	adminList, err := repo.List(context.Background(), 10, 0, "admin", true, false)
	if err != nil {
		t.Fatalf("admin list error: %v", err)
	}
	if len(adminList) != 3 {
		t.Fatalf("admin list len = %d, want 3", len(adminList))
	}

	overflow, err := repo.List(context.Background(), 10, 10, "", false, false)
	if err != nil {
		t.Fatalf("overflow list error: %v", err)
	}
	if len(overflow) != 0 {
		t.Fatalf("overflow len = %d, want 0", len(overflow))
	}
}

func TestEventRepository_List_FilterAndPaging(t *testing.T) {
	repo := NewEventRepository()
	now := time.Now()

	_ = repo.Create(context.Background(), &event.Event{
		ID:         "ev-draft-managed",
		Title:      "Draft Managed",
		Status:     event.EventStatusDraft,
		CreatorID:  "creator-1",
		ManagerIDs: []string{"manager-1"},
		CreatedAt:  now.Add(-1 * time.Hour),
	})
	_ = repo.Create(context.Background(), &event.Event{
		ID:        "ev-active-public",
		Title:     "Active Public",
		Status:    event.EventStatusActive,
		CreatedAt: now.Add(-2 * time.Hour),
	})
	_ = repo.Create(context.Background(), &event.Event{
		ID:        "ev-ended-public",
		Title:     "Ended Public",
		Status:    event.EventStatusEnded,
		CreatedAt: now.Add(-3 * time.Hour),
	})

	publicList, err := repo.List(context.Background(), 10, 0, "", false, false)
	if err != nil {
		t.Fatalf("public list error: %v", err)
	}
	if len(publicList) != 2 {
		t.Fatalf("public list len = %d, want 2", len(publicList))
	}
	if publicList[0].ID != "ev-active-public" {
		t.Fatalf("public list order[0] = %q, want %q", publicList[0].ID, "ev-active-public")
	}

	managerList, err := repo.List(context.Background(), 10, 0, "manager-1", false, true)
	if err != nil {
		t.Fatalf("manager list error: %v", err)
	}
	if len(managerList) != 1 || managerList[0].ID != "ev-draft-managed" {
		t.Fatalf("manager list mismatch: %+v", managerList)
	}

	adminList, err := repo.List(context.Background(), 10, 0, "admin", true, false)
	if err != nil {
		t.Fatalf("admin list error: %v", err)
	}
	if len(adminList) != 3 {
		t.Fatalf("admin list len = %d, want 3", len(adminList))
	}

	overflow, err := repo.List(context.Background(), 10, 10, "", false, false)
	if err != nil {
		t.Fatalf("overflow list error: %v", err)
	}
	if len(overflow) != 0 {
		t.Fatalf("overflow len = %d, want 0", len(overflow))
	}
}

func TestGroupBuyRepository_CreateOrder_RespectsStockLimit(t *testing.T) {
	repo := NewGroupBuyRepository()
	now := time.Now()
	_ = repo.Create(context.Background(), &groupbuy.GroupBuy{
		ID:     "gb-stock",
		Title:  "Stock Test",
		Status: groupbuy.GroupBuyStatusActive,
		Products: []*groupbuy.Product{
			{ID: "prod-1", MaxQuantity: 2},
		},
		CreatedAt: now,
	})

	err := repo.CreateOrder(context.Background(), &groupbuy.Order{
		ID:         "order-1",
		GroupBuyID: "gb-stock",
		UserID:     "u-1",
		Items: []*groupbuy.OrderItem{
			{ProductID: "prod-1", Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("first order should succeed: %v", err)
	}

	err = repo.CreateOrder(context.Background(), &groupbuy.Order{
		ID:         "order-2",
		GroupBuyID: "gb-stock",
		UserID:     "u-2",
		Items: []*groupbuy.OrderItem{
			{ProductID: "prod-1", Quantity: 2},
		},
	})
	if err == nil {
		t.Fatal("expected out of stock error")
	}
}

func TestEventRepository_Register_RespectsCapacity(t *testing.T) {
	repo := NewEventRepository()
	now := time.Now()
	_ = repo.Create(context.Background(), &event.Event{
		ID:     "ev-capacity",
		Title:  "Capacity Test",
		Status: event.EventStatusActive,
		Items: []*event.EventItem{
			{ID: "item-1", Name: "Workshop", MaxParticipants: 2},
		},
		CreatedAt: now,
	})

	err := repo.Register(context.Background(), &event.Registration{
		ID:      "reg-1",
		EventID: "ev-capacity",
		UserID:  "u-1",
		SelectedItems: []*event.RegistrationItem{
			{EventItemID: "item-1", Quantity: 1},
		},
		Status: event.RegistrationStatusPending,
	})
	if err != nil {
		t.Fatalf("first registration should succeed: %v", err)
	}

	err = repo.Register(context.Background(), &event.Registration{
		ID:      "reg-2",
		EventID: "ev-capacity",
		UserID:  "u-2",
		SelectedItems: []*event.RegistrationItem{
			{EventItemID: "item-1", Quantity: 2},
		},
		Status: event.RegistrationStatusPending,
	})
	if err == nil {
		t.Fatal("expected registration limit exceeded error")
	}
}
