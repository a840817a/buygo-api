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
		t.Fatal("got empty order id")
	}

	// Verify order persisted correctly by fetching it back
	orderResp, err := h.GetMyGroupBuyOrder(userCtx, connect.NewRequest(&v1.GetMyGroupBuyOrderRequest{
		GroupBuyId: gbID,
	}))
	if err != nil {
		t.Fatalf("GetMyGroupBuyOrder error: %v", err)
	}
	if orderResp.Msg.Order == nil {
		t.Fatal("expected order, got nil")
	}
	if orderResp.Msg.Order.GroupBuyId != gbID {
		t.Errorf("order group_buy_id = %q, want %q", orderResp.Msg.Order.GroupBuyId, gbID)
	}
	if len(orderResp.Msg.Order.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(orderResp.Msg.Order.Items))
	}
	if orderResp.Msg.Order.Items[0].Quantity != 2 {
		t.Errorf("item quantity = %d, want 2", orderResp.Msg.Order.Items[0].Quantity)
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

// createActiveGroupBuyWithOrder is a helper that creates an active group buy with a product and an order.
func createActiveGroupBuyWithOrder(t *testing.T, h *GroupBuyHandler) (gbID, prodID, orderID string) {
	t.Helper()
	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	createResp, err := h.CreateGroupBuy(creatorCtx, connect.NewRequest(&v1.CreateGroupBuyRequest{
		Title: "Test GB",
		Products: []*v1.Product{
			{Name: "Product 1", PriceOriginal: 100},
		},
	}))
	if err != nil {
		t.Fatalf("CreateGroupBuy: %v", err)
	}
	gbID = createResp.Msg.GroupBuy.Id

	_, err = h.UpdateGroupBuy(creatorCtx, connect.NewRequest(&v1.UpdateGroupBuyRequest{
		GroupBuyId: gbID,
		Status:     v1.GroupBuyStatus_GROUP_BUY_STATUS_ACTIVE,
		Products:   []*v1.Product{{Name: "Product 1", PriceOriginal: 100}},
	}))
	if err != nil {
		t.Fatalf("UpdateGroupBuy: %v", err)
	}

	getResp, _ := h.GetGroupBuy(context.Background(), connect.NewRequest(&v1.GetGroupBuyRequest{GroupBuyId: gbID}))
	prodID = getResp.Msg.Products[0].Id

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	orderResp, err := h.CreateOrder(userCtx, connect.NewRequest(&v1.CreateOrderRequest{
		GroupBuyId: gbID,
		Items:      []*v1.CreateOrderItem{{ProductId: prodID, Quantity: 2}},
	}))
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	orderID = orderResp.Msg.OrderId
	return
}

func TestGroupBuyHandler_ListGroupBuyOrders(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	gbID, _, _ := createActiveGroupBuyWithOrder(t, h)

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	resp, err := h.ListGroupBuyOrders(creatorCtx, connect.NewRequest(&v1.ListGroupBuyOrdersRequest{
		GroupBuyId: gbID,
	}))
	if err != nil {
		t.Fatalf("ListGroupBuyOrders error: %v", err)
	}
	if len(resp.Msg.Orders) != 1 {
		t.Errorf("got %d orders, want 1", len(resp.Msg.Orders))
	}
}

func TestGroupBuyHandler_GetMyOrders(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	createActiveGroupBuyWithOrder(t, h)

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	resp, err := h.GetMyOrders(userCtx, connect.NewRequest(&v1.GetMyOrdersRequest{}))
	if err != nil {
		t.Fatalf("GetMyOrders error: %v", err)
	}
	if len(resp.Msg.Orders) != 1 {
		t.Errorf("got %d orders, want 1", len(resp.Msg.Orders))
	}
}

func TestGroupBuyHandler_GetMyGroupBuyOrder(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	gbID, prodID, orderID := createActiveGroupBuyWithOrder(t, h)

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	resp, err := h.GetMyGroupBuyOrder(userCtx, connect.NewRequest(&v1.GetMyGroupBuyOrderRequest{
		GroupBuyId: gbID,
	}))
	if err != nil {
		t.Fatalf("GetMyGroupBuyOrder error: %v", err)
	}
	if resp.Msg.Order == nil {
		t.Fatal("expected order, got nil")
	}
	if resp.Msg.Order.Id != orderID {
		t.Errorf("order id = %q, want %q", resp.Msg.Order.Id, orderID)
	}
	if resp.Msg.Order.GroupBuyId != gbID {
		t.Errorf("group_buy_id = %q, want %q", resp.Msg.Order.GroupBuyId, gbID)
	}
	if len(resp.Msg.Order.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(resp.Msg.Order.Items))
	}
	if resp.Msg.Order.Items[0].ProductId != prodID {
		t.Errorf("item product_id = %q, want %q", resp.Msg.Order.Items[0].ProductId, prodID)
	}
}

func TestGroupBuyHandler_GetMyGroupBuyOrder_NoOrder(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	createResp, _ := h.CreateGroupBuy(creatorCtx, connect.NewRequest(&v1.CreateGroupBuyRequest{
		Title: "Empty GB",
	}))
	gbID := createResp.Msg.GroupBuy.Id

	userCtx := domainAuth.NewContext(context.Background(), "user-2", int(user.UserRoleUser))
	resp, err := h.GetMyGroupBuyOrder(userCtx, connect.NewRequest(&v1.GetMyGroupBuyOrderRequest{
		GroupBuyId: gbID,
	}))
	if err != nil {
		t.Fatalf("GetMyGroupBuyOrder error: %v", err)
	}
	if resp.Msg.Order != nil {
		t.Error("expected nil order for user with no order")
	}
}

func TestGroupBuyHandler_UpdateOrder(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	_, prodID, orderID := createActiveGroupBuyWithOrder(t, h)

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	resp, err := h.UpdateOrder(userCtx, connect.NewRequest(&v1.UpdateOrderRequest{
		OrderId: orderID,
		Items: []*v1.CreateOrderItem{
			{ProductId: prodID, Quantity: 5},
		},
		Note: "updated note",
	}))
	if err != nil {
		t.Fatalf("UpdateOrder error: %v", err)
	}
	if resp.Msg.Order == nil {
		t.Fatal("expected order in response")
	}
	if len(resp.Msg.Order.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(resp.Msg.Order.Items))
	}
	if resp.Msg.Order.Items[0].Quantity != 5 {
		t.Errorf("item quantity = %d, want 5", resp.Msg.Order.Items[0].Quantity)
	}
	if resp.Msg.Order.Note != "updated note" {
		t.Errorf("note = %q, want %q", resp.Msg.Order.Note, "updated note")
	}
}

func TestGroupBuyHandler_UpdatePaymentInfo(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	_, _, orderID := createActiveGroupBuyWithOrder(t, h)

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	paidAt := time.Now()
	resp, err := h.UpdatePaymentInfo(userCtx, connect.NewRequest(&v1.UpdatePaymentInfoRequest{
		OrderId:         orderID,
		Method:          "bank_transfer",
		AccountLast5:    "54321",
		ContactInfo:     "John",
		ShippingAddress: "123 Main St",
		PaidAt:          timestamppb.New(paidAt),
		Amount:          200,
	}))
	if err != nil {
		t.Fatalf("UpdatePaymentInfo error: %v", err)
	}
	if resp.Msg.Order == nil {
		t.Fatal("expected order in response")
	}
	if resp.Msg.Order.PaymentInfo == nil {
		t.Fatal("expected payment_info in order")
	}
	if resp.Msg.Order.PaymentInfo.Method != "bank_transfer" {
		t.Errorf("method = %q, want %q", resp.Msg.Order.PaymentInfo.Method, "bank_transfer")
	}
	if resp.Msg.Order.PaymentInfo.AccountLast5 != "54321" {
		t.Errorf("account_last5 = %q, want %q", resp.Msg.Order.PaymentInfo.AccountLast5, "54321")
	}
	if resp.Msg.Order.ContactInfo != "John" {
		t.Errorf("contact_info = %q, want %q", resp.Msg.Order.ContactInfo, "John")
	}
	if resp.Msg.Order.ShippingAddress != "123 Main St" {
		t.Errorf("shipping_address = %q, want %q", resp.Msg.Order.ShippingAddress, "123 Main St")
	}
	if resp.Msg.Order.PaymentInfo.Amount != 200 {
		t.Errorf("amount = %d, want 200", resp.Msg.Order.PaymentInfo.Amount)
	}
}

func TestGroupBuyHandler_ConfirmPayment(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	gbID, _, orderID := createActiveGroupBuyWithOrder(t, h)

	// First submit payment info as user
	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	_, err := h.UpdatePaymentInfo(userCtx, connect.NewRequest(&v1.UpdatePaymentInfoRequest{
		OrderId:      orderID,
		Method:       "bank_transfer",
		AccountLast5: "54321",
		Amount:       200,
	}))
	if err != nil {
		t.Fatalf("UpdatePaymentInfo error: %v", err)
	}

	// Confirm as creator/manager
	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	resp, err := h.ConfirmPayment(creatorCtx, connect.NewRequest(&v1.ConfirmPaymentRequest{
		OrderId: orderID,
		Status:  v1.PaymentStatus_PAYMENT_STATUS_CONFIRMED,
	}))
	if err != nil {
		t.Fatalf("ConfirmPayment error: %v", err)
	}
	if resp.Msg.OrderId != orderID {
		t.Errorf("got order id %q, want %q", resp.Msg.OrderId, orderID)
	}
	if resp.Msg.Status != v1.PaymentStatus_PAYMENT_STATUS_CONFIRMED {
		t.Errorf("got status %v, want PAYMENT_STATUS_CONFIRMED", resp.Msg.Status)
	}

	// Verify status persisted by fetching the order
	orderResp, err := h.GetMyGroupBuyOrder(userCtx, connect.NewRequest(&v1.GetMyGroupBuyOrderRequest{
		GroupBuyId: gbID,
	}))
	if err != nil {
		t.Fatalf("GetMyGroupBuyOrder error: %v", err)
	}
	if orderResp.Msg.Order.PaymentStatus != v1.PaymentStatus_PAYMENT_STATUS_CONFIRMED {
		t.Errorf("persisted payment status = %v, want CONFIRMED", orderResp.Msg.Order.PaymentStatus)
	}
}

func TestGroupBuyHandler_CancelOrder(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	gbID, _, orderID := createActiveGroupBuyWithOrder(t, h)

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	resp, err := h.CancelOrder(userCtx, connect.NewRequest(&v1.CancelOrderRequest{
		OrderId: orderID,
	}))
	if err != nil {
		t.Fatalf("CancelOrder error: %v", err)
	}
	if resp.Msg.OrderId != orderID {
		t.Errorf("got order id %q, want %q", resp.Msg.OrderId, orderID)
	}

	// Verify order is actually cancelled by fetching it
	orderResp, err := h.GetMyGroupBuyOrder(userCtx, connect.NewRequest(&v1.GetMyGroupBuyOrderRequest{
		GroupBuyId: gbID,
	}))
	if err != nil {
		t.Fatalf("GetMyGroupBuyOrder error: %v", err)
	}
	if orderResp.Msg.Order == nil {
		t.Fatal("expected order, got nil")
	}
	if orderResp.Msg.Order.PaymentStatus != v1.PaymentStatus_PAYMENT_STATUS_REJECTED {
		t.Errorf("order payment_status = %v, want REJECTED (cancelled)", orderResp.Msg.Order.PaymentStatus)
	}
}

func TestGroupBuyHandler_CreateCategory(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	ctx := domainAuth.NewContext(context.Background(), "admin-1", int(user.UserRoleSysAdmin))
	resp, err := h.CreateCategory(ctx, connect.NewRequest(&v1.CreateCategoryRequest{
		Name:      "Electronics",
		SpecNames: []string{"Color", "Size"},
	}))
	if err != nil {
		t.Fatalf("CreateCategory error: %v", err)
	}
	if resp.Msg.Category.Name != "Electronics" {
		t.Errorf("got name %q, want %q", resp.Msg.Category.Name, "Electronics")
	}
}

func TestGroupBuyHandler_ListCategories(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	ctx := domainAuth.NewContext(context.Background(), "admin-1", int(user.UserRoleSysAdmin))
	if _, err := h.CreateCategory(ctx, connect.NewRequest(&v1.CreateCategoryRequest{
		Name: "Cat 1",
	})); err != nil {
		t.Fatalf("CreateCategory(Cat 1): %v", err)
	}
	if _, err := h.CreateCategory(ctx, connect.NewRequest(&v1.CreateCategoryRequest{
		Name: "Cat 2",
	})); err != nil {
		t.Fatalf("CreateCategory(Cat 2): %v", err)
	}

	resp, err := h.ListCategories(ctx, connect.NewRequest(&v1.ListCategoriesRequest{}))
	if err != nil {
		t.Fatalf("ListCategories error: %v", err)
	}
	if len(resp.Msg.Categories) != 2 {
		t.Errorf("got %d categories, want 2", len(resp.Msg.Categories))
	}
}

func TestGroupBuyHandler_PriceTemplateCRUD(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	ctx := domainAuth.NewContext(context.Background(), "admin-1", int(user.UserRoleSysAdmin))

	// Create
	createResp, err := h.CreatePriceTemplate(ctx, connect.NewRequest(&v1.CreatePriceTemplateRequest{
		Name:           "JPY Template",
		SourceCurrency: "JPY",
		ExchangeRate:   0.22,
		RoundingConfig: &v1.RoundingConfig{
			Method: v1.RoundingMethod_ROUNDING_METHOD_ROUND,
			Digit:  0,
		},
	}))
	if err != nil {
		t.Fatalf("CreatePriceTemplate error: %v", err)
	}
	ptID := createResp.Msg.Template.Id
	if createResp.Msg.Template.Name != "JPY Template" {
		t.Errorf("got name %q, want %q", createResp.Msg.Template.Name, "JPY Template")
	}

	// Get
	getResp, err := h.GetPriceTemplate(ctx, connect.NewRequest(&v1.GetPriceTemplateRequest{
		TemplateId: ptID,
	}))
	if err != nil {
		t.Fatalf("GetPriceTemplate error: %v", err)
	}
	if getResp.Msg.Template.Id != ptID {
		t.Errorf("got id %q, want %q", getResp.Msg.Template.Id, ptID)
	}

	// List
	listResp, err := h.ListPriceTemplates(ctx, connect.NewRequest(&v1.ListPriceTemplatesRequest{}))
	if err != nil {
		t.Fatalf("ListPriceTemplates error: %v", err)
	}
	if len(listResp.Msg.Templates) != 1 {
		t.Errorf("got %d templates, want 1", len(listResp.Msg.Templates))
	}

	// Update
	updateResp, err := h.UpdatePriceTemplate(ctx, connect.NewRequest(&v1.UpdatePriceTemplateRequest{
		TemplateId:     ptID,
		Name:           "USD Template",
		SourceCurrency: "USD",
		ExchangeRate:   31.5,
	}))
	if err != nil {
		t.Fatalf("UpdatePriceTemplate error: %v", err)
	}
	if updateResp.Msg.Template.Name != "USD Template" {
		t.Errorf("got name %q, want %q", updateResp.Msg.Template.Name, "USD Template")
	}

	// Delete
	deleteResp, err := h.DeletePriceTemplate(ctx, connect.NewRequest(&v1.DeletePriceTemplateRequest{
		TemplateId: ptID,
	}))
	if err != nil {
		t.Fatalf("DeletePriceTemplate error: %v", err)
	}
	if deleteResp.Msg.TemplateId != ptID {
		t.Errorf("got id %q, want %q", deleteResp.Msg.TemplateId, ptID)
	}

	// Verify deleted
	listResp2, _ := h.ListPriceTemplates(ctx, connect.NewRequest(&v1.ListPriceTemplatesRequest{}))
	if len(listResp2.Msg.Templates) != 0 {
		t.Errorf("got %d templates after delete, want 0", len(listResp2.Msg.Templates))
	}
}

func TestGroupBuyHandler_ListGroupBuys(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	ctx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	resp1, _ := h.CreateGroupBuy(ctx, connect.NewRequest(&v1.CreateGroupBuyRequest{Title: "GB 1"}))
	resp2, _ := h.CreateGroupBuy(ctx, connect.NewRequest(&v1.CreateGroupBuyRequest{Title: "GB 2"}))

	// Activate both so they appear in public list
	_, _ = h.UpdateGroupBuy(ctx, connect.NewRequest(&v1.UpdateGroupBuyRequest{
		GroupBuyId: resp1.Msg.GroupBuy.Id,
		Status:     v1.GroupBuyStatus_GROUP_BUY_STATUS_ACTIVE,
	}))
	_, _ = h.UpdateGroupBuy(ctx, connect.NewRequest(&v1.UpdateGroupBuyRequest{
		GroupBuyId: resp2.Msg.GroupBuy.Id,
		Status:     v1.GroupBuyStatus_GROUP_BUY_STATUS_ACTIVE,
	}))

	resp, err := h.ListGroupBuys(ctx, connect.NewRequest(&v1.ListGroupBuysRequest{
		PageSize: 10,
	}))
	if err != nil {
		t.Fatalf("ListGroupBuys error: %v", err)
	}
	if len(resp.Msg.GroupBuys) < 2 {
		t.Errorf("got %d group buys, want at least 2", len(resp.Msg.GroupBuys))
	}
}

func TestGroupBuyHandler_CreateOrder_Validation(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	ctx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))

	// Empty group_buy_id
	_, err := h.CreateOrder(ctx, connect.NewRequest(&v1.CreateOrderRequest{
		Items: []*v1.CreateOrderItem{{ProductId: "p1", Quantity: 1}},
	}))
	if err == nil {
		t.Fatal("expected error for empty group_buy_id")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("empty group_buy_id: got code %v, want CodeInvalidArgument", connect.CodeOf(err))
	}

	// No items
	_, err = h.CreateOrder(ctx, connect.NewRequest(&v1.CreateOrderRequest{
		GroupBuyId: "gb1",
	}))
	if err == nil {
		t.Fatal("expected error for empty items")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("empty items: got code %v, want CodeInvalidArgument", connect.CodeOf(err))
	}

	// Zero quantity
	_, err = h.CreateOrder(ctx, connect.NewRequest(&v1.CreateOrderRequest{
		GroupBuyId: "gb1",
		Items:     []*v1.CreateOrderItem{{ProductId: "p1", Quantity: 0}},
	}))
	if err == nil {
		t.Fatal("expected error for zero quantity")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("zero quantity: got code %v, want CodeInvalidArgument", connect.CodeOf(err))
	}
}
