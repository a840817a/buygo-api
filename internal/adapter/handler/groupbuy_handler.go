package handler

import (
	"context"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/buygo/buygo-api/api/v1"
	"github.com/buygo/buygo-api/api/v1/buygov1connect"
	"github.com/buygo/buygo-api/internal/domain/groupbuy"
	"github.com/buygo/buygo-api/internal/service"
)

type GroupBuyHandler struct {
	svc *service.GroupBuyService
}

func NewGroupBuyHandler(svc *service.GroupBuyService) *GroupBuyHandler {
	return &GroupBuyHandler{svc: svc}
}

// Ensure implementation
var _ buygov1connect.GroupBuyServiceHandler = (*GroupBuyHandler)(nil)

func (h *GroupBuyHandler) CreateGroupBuy(ctx context.Context, req *connect.Request[v1.CreateGroupBuyRequest]) (*connect.Response[v1.CreateGroupBuyResponse], error) {
	gb, err := h.svc.CreateGroupBuy(ctx, req.Msg.Title, req.Msg.Description)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.CreateGroupBuyResponse{
		GroupBuy: toProtoGroupBuy(gb),
	}), nil
}

func (h *GroupBuyHandler) ListGroupBuys(ctx context.Context, req *connect.Request[v1.ListGroupBuysRequest]) (*connect.Response[v1.ListGroupBuysResponse], error) {
	gbs, err := h.svc.ListGroupBuys(ctx, int(req.Msg.PageSize), 0)
	if err != nil {
		return nil, mapError(err)
	}

	var protoGroupBuys []*v1.GroupBuy
	for _, gb := range gbs {
		protoGroupBuys = append(protoGroupBuys, toProtoGroupBuy(gb))
	}

	return connect.NewResponse(&v1.ListGroupBuysResponse{
		GroupBuys: protoGroupBuys,
	}), nil
}

func (h *GroupBuyHandler) ListManagerGroupBuys(ctx context.Context, req *connect.Request[v1.ListManagerGroupBuysRequest]) (*connect.Response[v1.ListManagerGroupBuysResponse], error) {
	gbs, err := h.svc.ListManagerGroupBuys(ctx, int(req.Msg.PageSize), 0)
	if err != nil {
		return nil, mapError(err)
	}

	var protoGroupBuys []*v1.GroupBuy
	for _, gb := range gbs {
		protoGroupBuys = append(protoGroupBuys, toProtoGroupBuy(gb))
	}

	return connect.NewResponse(&v1.ListManagerGroupBuysResponse{
		GroupBuys: protoGroupBuys,
	}), nil
}

func (h *GroupBuyHandler) GetGroupBuy(ctx context.Context, req *connect.Request[v1.GetGroupBuyRequest]) (*connect.Response[v1.GetGroupBuyResponse], error) {
	gb, err := h.svc.GetGroupBuy(ctx, req.Msg.GroupBuyId)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.GetGroupBuyResponse{
		GroupBuy: toProtoGroupBuy(gb),
		Products: toProtoProducts(gb.Products),
	}), nil
}

func (h *GroupBuyHandler) UpdateGroupBuy(ctx context.Context, req *connect.Request[v1.UpdateGroupBuyRequest]) (*connect.Response[v1.UpdateGroupBuyResponse], error) {
	// Use make to ensure non-nil slices, allowing empty lists to clear data
	status := groupbuy.GroupBuyStatus(req.Msg.Status)

	products := make([]*groupbuy.Product, 0, len(req.Msg.Products))
	for _, p := range req.Msg.Products {
		products = append(products, fromProtoProduct(p))
	}

	var deadline *time.Time
	if req.Msg.Deadline != nil {
		t := req.Msg.Deadline.AsTime()
		deadline = &t
	}

	shippingConfigs := make([]*groupbuy.ShippingConfig, 0, len(req.Msg.ShippingConfigs))
	for _, sc := range req.Msg.ShippingConfigs {
		shippingConfigs = append(shippingConfigs, fromProtoShippingConfig(sc))
	}

	var rounding *groupbuy.RoundingConfig
	if req.Msg.RoundingConfig != nil {
		rounding = &groupbuy.RoundingConfig{
			Method: groupbuy.RoundingMethod(req.Msg.RoundingConfig.Method),
			Digit:  int(req.Msg.RoundingConfig.Digit),
		}
	}

	gb, err := h.svc.UpdateGroupBuy(ctx, req.Msg.GroupBuyId, req.Msg.Title, req.Msg.Description, status, products, req.Msg.CoverImageUrl, deadline, shippingConfigs, req.Msg.ManagerIds, req.Msg.ExchangeRate, rounding, req.Msg.SourceCurrency)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.UpdateGroupBuyResponse{
		GroupBuy: toProtoGroupBuy(gb),
	}), nil
}

func (h *GroupBuyHandler) AddProduct(ctx context.Context, req *connect.Request[v1.AddProductRequest]) (*connect.Response[v1.AddProductResponse], error) {
	p, err := h.svc.AddProduct(ctx, req.Msg.GroupBuyId, req.Msg.Name, req.Msg.PriceOriginal, req.Msg.ExchangeRate, req.Msg.Specs)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.AddProductResponse{
		Product: toProtoProduct(p),
	}), nil
}

func (h *GroupBuyHandler) CreateOrder(ctx context.Context, req *connect.Request[v1.CreateOrderRequest]) (*connect.Response[v1.CreateOrderResponse], error) {
	// Map Request Items to Domain Items
	var items []*groupbuy.OrderItem
	for _, i := range req.Msg.Items {
		items = append(items, &groupbuy.OrderItem{
			ProductID: i.ProductId,
			SpecID:    i.SpecId,
			Quantity:  int(i.Quantity),
		})
	}

	order, err := h.svc.CreateOrder(ctx, req.Msg.GroupBuyId, items, req.Msg.ContactInfo, req.Msg.ShippingAddress, req.Msg.ShippingMethodId, req.Msg.Note)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.CreateOrderResponse{
		OrderId: order.ID,
	}), nil
}

func (h *GroupBuyHandler) ListGroupBuyOrders(ctx context.Context, req *connect.Request[v1.ListGroupBuyOrdersRequest]) (*connect.Response[v1.ListGroupBuyOrdersResponse], error) {
	orders, err := h.svc.ListGroupBuyOrders(ctx, req.Msg.GroupBuyId)
	if err != nil {
		return nil, mapError(err)
	}

	var protoOrders []*v1.Order
	for _, o := range orders {
		protoOrders = append(protoOrders, toProtoOrder(o))
	}

	return connect.NewResponse(&v1.ListGroupBuyOrdersResponse{
		Orders: protoOrders,
	}), nil
}

func (h *GroupBuyHandler) GetMyOrders(ctx context.Context, req *connect.Request[v1.GetMyOrdersRequest]) (*connect.Response[v1.GetMyOrdersResponse], error) {
	orders, err := h.svc.GetMyOrders(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	// Map to Proto
	var protoOrders []*v1.Order
	for _, o := range orders {
		protoOrders = append(protoOrders, toProtoOrder(o))
	}

	return connect.NewResponse(&v1.GetMyOrdersResponse{
		Orders: protoOrders,
	}), nil
}

func (h *GroupBuyHandler) BatchUpdateStatus(ctx context.Context, req *connect.Request[v1.BatchUpdateStatusRequest]) (*connect.Response[v1.BatchUpdateStatusResponse], error) {
	updatedCount, updatedIds, err := h.svc.BatchUpdateStatus(ctx, req.Msg.GroupBuyId, req.Msg.SpecId, int(req.Msg.TargetStatus), req.Msg.Count)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.BatchUpdateStatusResponse{
		UpdatedCount:    updatedCount,
		UpdatedOrderIds: updatedIds,
	}), nil
}

func (h *GroupBuyHandler) ConfirmPayment(ctx context.Context, req *connect.Request[v1.ConfirmPaymentRequest]) (*connect.Response[v1.ConfirmPaymentResponse], error) {
	status := groupbuy.PaymentStatus(req.Msg.Status)
	if err := h.svc.ConfirmPayment(ctx, req.Msg.OrderId, status); err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.ConfirmPaymentResponse{
		OrderId: req.Msg.OrderId,
		Status:  req.Msg.Status,
	}), nil
}

func (h *GroupBuyHandler) GetMyGroupBuyOrder(ctx context.Context, req *connect.Request[v1.GetMyGroupBuyOrderRequest]) (*connect.Response[v1.GetMyGroupBuyOrderResponse], error) {
	order, err := h.svc.GetMyGroupBuyOrder(ctx, req.Msg.GroupBuyId)
	if err != nil {
		return nil, mapError(err)
	}

	var protoOrder *v1.Order
	if order != nil {
		protoOrder = toProtoOrder(order)
	}

	return connect.NewResponse(&v1.GetMyGroupBuyOrderResponse{
		Order: protoOrder,
	}), nil
}

func (h *GroupBuyHandler) UpdateOrder(ctx context.Context, req *connect.Request[v1.UpdateOrderRequest]) (*connect.Response[v1.UpdateOrderResponse], error) {
	var items []*groupbuy.OrderItem
	for _, i := range req.Msg.Items {
		items = append(items, &groupbuy.OrderItem{
			ProductID: i.ProductId,
			SpecID:    i.SpecId,
			Quantity:  int(i.Quantity),
			Status:    groupbuy.OrderItemStatus(i.Status),
		})
	}

	order, err := h.svc.UpdateOrder(ctx, req.Msg.OrderId, items, req.Msg.Note)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.UpdateOrderResponse{
		Order: toProtoOrder(order),
	}), nil
}

func (h *GroupBuyHandler) UpdatePaymentInfo(ctx context.Context, req *connect.Request[v1.UpdatePaymentInfoRequest]) (*connect.Response[v1.UpdatePaymentInfoResponse], error) {
	var paidAt *time.Time
	if req.Msg.PaidAt != nil {
		t := req.Msg.PaidAt.AsTime()
		paidAt = &t
	}
	order, err := h.svc.UpdatePaymentInfo(ctx, req.Msg.OrderId, req.Msg.Method, req.Msg.AccountLast5, req.Msg.ContactInfo, req.Msg.ShippingAddress, paidAt, req.Msg.Amount)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.UpdatePaymentInfoResponse{
		Order: toProtoOrder(order),
	}), nil
}

func (h *GroupBuyHandler) CancelOrder(ctx context.Context, req *connect.Request[v1.CancelOrderRequest]) (*connect.Response[v1.CancelOrderResponse], error) {
	if err := h.svc.CancelOrder(ctx, req.Msg.OrderId); err != nil {
		return nil, mapError(err)
	}
	return connect.NewResponse(&v1.CancelOrderResponse{
		OrderId: req.Msg.OrderId,
		// Status: ...
	}), nil
}

func (h *GroupBuyHandler) CreateCategory(ctx context.Context, req *connect.Request[v1.CreateCategoryRequest]) (*connect.Response[v1.CreateCategoryResponse], error) {
	c, err := h.svc.CreateCategory(ctx, req.Msg.Name, req.Msg.SpecNames)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.CreateCategoryResponse{
		Category: toProtoCategory(c),
	}), nil
}

func (h *GroupBuyHandler) ListCategories(ctx context.Context, req *connect.Request[v1.ListCategoriesRequest]) (*connect.Response[v1.ListCategoriesResponse], error) {
	categories, err := h.svc.ListCategories(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	var protoCategories []*v1.Category
	for _, c := range categories {
		protoCategories = append(protoCategories, toProtoCategory(c))
	}

	return connect.NewResponse(&v1.ListCategoriesResponse{
		Categories: protoCategories,
	}), nil
}

// PriceTemplate Handlers

func (h *GroupBuyHandler) CreatePriceTemplate(ctx context.Context, req *connect.Request[v1.CreatePriceTemplateRequest]) (*connect.Response[v1.CreatePriceTemplateResponse], error) {
	var rounding *groupbuy.RoundingConfig
	if req.Msg.RoundingConfig != nil {
		rounding = &groupbuy.RoundingConfig{
			Method: groupbuy.RoundingMethod(req.Msg.RoundingConfig.Method),
			Digit:  int(req.Msg.RoundingConfig.Digit),
		}
	}

	pt, err := h.svc.CreatePriceTemplate(ctx, req.Msg.Name, req.Msg.SourceCurrency, req.Msg.ExchangeRate, rounding)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.CreatePriceTemplateResponse{
		Template: toProtoPriceTemplate(pt),
	}), nil
}

func (h *GroupBuyHandler) ListPriceTemplates(ctx context.Context, req *connect.Request[v1.ListPriceTemplatesRequest]) (*connect.Response[v1.ListPriceTemplatesResponse], error) {
	templates, err := h.svc.ListPriceTemplates(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	var protoTemplates []*v1.PriceTemplate
	for _, pt := range templates {
		protoTemplates = append(protoTemplates, toProtoPriceTemplate(pt))
	}

	return connect.NewResponse(&v1.ListPriceTemplatesResponse{
		Templates: protoTemplates,
	}), nil
}

func (h *GroupBuyHandler) GetPriceTemplate(ctx context.Context, req *connect.Request[v1.GetPriceTemplateRequest]) (*connect.Response[v1.GetPriceTemplateResponse], error) {
	pt, err := h.svc.GetPriceTemplate(ctx, req.Msg.TemplateId)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.GetPriceTemplateResponse{
		Template: toProtoPriceTemplate(pt),
	}), nil
}

func (h *GroupBuyHandler) UpdatePriceTemplate(ctx context.Context, req *connect.Request[v1.UpdatePriceTemplateRequest]) (*connect.Response[v1.UpdatePriceTemplateResponse], error) {
	var rounding *groupbuy.RoundingConfig
	if req.Msg.RoundingConfig != nil {
		rounding = &groupbuy.RoundingConfig{
			Method: groupbuy.RoundingMethod(req.Msg.RoundingConfig.Method),
			Digit:  int(req.Msg.RoundingConfig.Digit),
		}
	}

	pt, err := h.svc.UpdatePriceTemplate(ctx, req.Msg.TemplateId, req.Msg.Name, req.Msg.SourceCurrency, req.Msg.ExchangeRate, rounding)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.UpdatePriceTemplateResponse{
		Template: toProtoPriceTemplate(pt),
	}), nil
}

func (h *GroupBuyHandler) DeletePriceTemplate(ctx context.Context, req *connect.Request[v1.DeletePriceTemplateRequest]) (*connect.Response[v1.DeletePriceTemplateResponse], error) {
	if err := h.svc.DeletePriceTemplate(ctx, req.Msg.TemplateId); err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.DeletePriceTemplateResponse{
		TemplateId: req.Msg.TemplateId,
	}), nil
}
