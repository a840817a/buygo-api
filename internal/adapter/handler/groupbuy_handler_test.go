package handler

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/hatsubosi/buygo-api/api/v1"
	"github.com/hatsubosi/buygo-api/internal/adapter/repository/memory"
	domainAuth "github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"github.com/hatsubosi/buygo-api/internal/service"
)

func TestGroupBuyHandler_CreateGroupBuy(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	ctx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	deadline := time.Now().Add(24 * time.Hour)

	req := &v1.CreateGroupBuyRequest{
		Title:       "Test Group Buy",
		Description: "A test group buy",
		Deadline:    timestamppb.New(deadline),
		Products: []*v1.Product{
			{
				Name:          "Product 1",
				Description:   "Desc 1",
				PriceOriginal: 100,
			},
		},
		ManagerIds: []string{"manager-1"},
	}

	resp, err := h.CreateGroupBuy(ctx, connect.NewRequest(req))
	if err != nil {
		t.Fatalf("CreateGroupBuy error: %v", err)
	}

	if resp.Msg.GroupBuy.Title != req.Title {
		t.Errorf("got title %q, want %q", resp.Msg.GroupBuy.Title, req.Title)
	}
}

func TestGroupBuyHandler_GetGroupBuy(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	ctx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	createResp, _ := h.CreateGroupBuy(ctx, connect.NewRequest(&v1.CreateGroupBuyRequest{
		Title: "Test GB",
	}))

	gbID := createResp.Msg.GroupBuy.Id

	getResp, err := h.GetGroupBuy(context.Background(), connect.NewRequest(&v1.GetGroupBuyRequest{
		GroupBuyId: gbID,
	}))
	if err != nil {
		t.Fatalf("GetGroupBuy error: %v", err)
	}

	if getResp.Msg.GroupBuy.Id != gbID {
		t.Errorf("got id %q, want %q", getResp.Msg.GroupBuy.Id, gbID)
	}
}

func TestGroupBuyHandler_AddProduct(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	ctx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	createResp, _ := h.CreateGroupBuy(ctx, connect.NewRequest(&v1.CreateGroupBuyRequest{
		Title: "Test GB",
	}))

	gbID := createResp.Msg.GroupBuy.Id

	_, err := h.AddProduct(ctx, connect.NewRequest(&v1.AddProductRequest{
		GroupBuyId:    gbID,
		Name:          "New Product",
		PriceOriginal: 200,
	}))
	if err != nil {
		t.Fatalf("AddProduct error: %v", err)
	}

	getResp, _ := h.GetGroupBuy(context.Background(), connect.NewRequest(&v1.GetGroupBuyRequest{
		GroupBuyId: gbID,
	}))

	found := false
	for _, p := range getResp.Msg.Products {
		if p.Name == "New Product" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("added product not found in GroupBuy")
	}
}

func TestGroupBuyHandler_CreateOrder(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	createResp, _ := h.CreateGroupBuy(creatorCtx, connect.NewRequest(&v1.CreateGroupBuyRequest{
		Title: "Test GB",
		Products: []*v1.Product{
			{Name: "Product 1", PriceOriginal: 100},
		},
	}))

	gbID := createResp.Msg.GroupBuy.Id

	// Activate GroupBuy and keep products
	_, _ = h.UpdateGroupBuy(creatorCtx, connect.NewRequest(&v1.UpdateGroupBuyRequest{
		GroupBuyId: gbID,
		Status:     v1.GroupBuyStatus_GROUP_BUY_STATUS_ACTIVE,
		Products: []*v1.Product{
			{Name: "Product 1", PriceOriginal: 100},
		},
	}))

	// Get products to get the ID
	getResp, _ := h.GetGroupBuy(context.Background(), connect.NewRequest(&v1.GetGroupBuyRequest{GroupBuyId: gbID}))
	prodID := getResp.Msg.Products[0].Id

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	resp, err := h.CreateOrder(userCtx, connect.NewRequest(&v1.CreateOrderRequest{
		GroupBuyId: gbID,
		Items: []*v1.CreateOrderItem{
			{ProductId: prodID, Quantity: 2},
		},
	}))
	if err != nil {
		t.Fatalf("CreateOrder error: %v", err)
	}

	if resp.Msg.OrderId == "" {
		t.Errorf("got empty order id")
	}
}

func TestGroupBuyHandler_ListManagerGroupBuys(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	managerCtx := domainAuth.NewContext(context.Background(), "manager-1", int(user.UserRoleCreator))
	h.CreateGroupBuy(managerCtx, connect.NewRequest(&v1.CreateGroupBuyRequest{
		Title:      "GB 1",
		ManagerIds: []string{"manager-1"},
	}))

	resp, err := h.ListManagerGroupBuys(managerCtx, connect.NewRequest(&v1.ListManagerGroupBuysRequest{
		PageSize: 10,
	}))
	if err != nil {
		t.Fatalf("ListManagerGroupBuys error: %v", err)
	}

	if len(resp.Msg.GroupBuys) != 1 {
		t.Errorf("got %d group buys, want 1", len(resp.Msg.GroupBuys))
	}
}
