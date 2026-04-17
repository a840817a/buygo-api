package handler

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/buygo/buygo-api/api/v1"
	"github.com/buygo/buygo-api/api/v1/buygov1connect"
	"github.com/buygo/buygo-api/internal/domain/project"
	"github.com/buygo/buygo-api/internal/service"
)

type ProjectHandler struct {
	svc *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// Ensure implementation
var _ buygov1connect.ProjectServiceHandler = (*ProjectHandler)(nil)

func (h *ProjectHandler) CreateProject(ctx context.Context, req *connect.Request[v1.CreateProjectRequest]) (*connect.Response[v1.CreateProjectResponse], error) {
	p, err := h.svc.CreateProject(ctx, req.Msg.Title, req.Msg.Description)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.CreateProjectResponse{
		Project: toProtoProject(p),
	}), nil
}

func (h *ProjectHandler) ListProjects(ctx context.Context, req *connect.Request[v1.ListProjectsRequest]) (*connect.Response[v1.ListProjectsResponse], error) {
	// TODO: Implement List in Service
	projects, err := h.svc.ListProjects(ctx, int(req.Msg.PageSize), 0)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var protoProjects []*v1.Project
	for _, p := range projects {
		protoProjects = append(protoProjects, toProtoProject(p))
	}

	return connect.NewResponse(&v1.ListProjectsResponse{
		Projects: protoProjects,
	}), nil
}

func (h *ProjectHandler) ListManagerProjects(ctx context.Context, req *connect.Request[v1.ListManagerProjectsRequest]) (*connect.Response[v1.ListManagerProjectsResponse], error) {
	projects, err := h.svc.ListManagerProjects(ctx, int(req.Msg.PageSize), 0)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var protoProjects []*v1.Project
	for _, p := range projects {
		protoProjects = append(protoProjects, toProtoProject(p))
	}

	return connect.NewResponse(&v1.ListManagerProjectsResponse{
		Projects: protoProjects,
	}), nil
}

func (h *ProjectHandler) GetProject(ctx context.Context, req *connect.Request[v1.GetProjectRequest]) (*connect.Response[v1.GetProjectResponse], error) {
	p, err := h.svc.GetProject(ctx, req.Msg.ProjectId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&v1.GetProjectResponse{
		Project:  toProtoProject(p),
		Products: toProtoProducts(p.Products),
	}), nil
}

func (h *ProjectHandler) UpdateProject(ctx context.Context, req *connect.Request[v1.UpdateProjectRequest]) (*connect.Response[v1.UpdateProjectResponse], error) {
	// Use make to ensure non-nil slices, allowing empty lists to clear data
	status := project.ProjectStatus(req.Msg.Status)

	products := make([]*project.Product, 0, len(req.Msg.Products))
	for _, p := range req.Msg.Products {
		products = append(products, fromProtoProduct(p))
	}

	var deadline *time.Time
	if req.Msg.Deadline != nil {
		t := req.Msg.Deadline.AsTime()
		deadline = &t
	}

	shippingConfigs := make([]*project.ShippingConfig, 0, len(req.Msg.ShippingConfigs))
	for _, sc := range req.Msg.ShippingConfigs {
		shippingConfigs = append(shippingConfigs, fromProtoShippingConfig(sc))
	}

	var rounding *project.RoundingConfig
	if req.Msg.RoundingConfig != nil {
		rounding = &project.RoundingConfig{
			Method: int(req.Msg.RoundingConfig.Method),
			Digit:  int(req.Msg.RoundingConfig.Digit),
		}
	}

	p, err := h.svc.UpdateProject(ctx, req.Msg.ProjectId, req.Msg.Title, req.Msg.Description, status, products, req.Msg.CoverImageUrl, deadline, shippingConfigs, req.Msg.ManagerIds, req.Msg.ExchangeRate, rounding, req.Msg.SourceCurrency)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.UpdateProjectResponse{
		Project: toProtoProject(p),
	}), nil
}

// Helpers

func fromProtoProduct(p *v1.Product) *project.Product {
	if p == nil {
		return nil
	}

	var specs []*project.ProductSpec
	for _, s := range p.Specs {
		specs = append(specs, &project.ProductSpec{
			ID:   s.Id,
			Name: s.Name,
		})
	}

	var rc *project.RoundingConfig
	if p.RoundingConfig != nil {
		rc = &project.RoundingConfig{
			Method: int(p.RoundingConfig.Method),
			Digit:  int(p.RoundingConfig.Digit),
		}
	}

	return &project.Product{
		ID:            p.Id,
		ProjectID:     p.ProjectId,
		Name:          p.Name,
		Description:   p.Description,
		ImageURL:      p.ImageUrl,
		PriceOriginal: p.PriceOriginal,
		ExchangeRate:  p.ExchangeRate,
		Rounding:      rc,
		PriceFinal:    p.PriceFinal,
		MaxQuantity:   p.MaxQuantity,
		Specs:         specs,
	}
}

func (h *ProjectHandler) AddProduct(ctx context.Context, req *connect.Request[v1.AddProductRequest]) (*connect.Response[v1.AddProductResponse], error) {
	p, err := h.svc.AddProduct(ctx, req.Msg.ProjectId, req.Msg.Name, req.Msg.PriceOriginal, req.Msg.ExchangeRate, req.Msg.Specs)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.AddProductResponse{
		Product: toProtoProduct(p),
	}), nil
}

func (h *ProjectHandler) CreateOrder(ctx context.Context, req *connect.Request[v1.CreateOrderRequest]) (*connect.Response[v1.CreateOrderResponse], error) {
	// Map Request Items to Domain Items
	var items []*project.OrderItem
	for _, i := range req.Msg.Items {
		items = append(items, &project.OrderItem{
			ProductID: i.ProductId,
			SpecID:    i.SpecId,
			Quantity:  int(i.Quantity),
		})
	}

	order, err := h.svc.CreateOrder(ctx, req.Msg.ProjectId, items, req.Msg.ContactInfo, req.Msg.ShippingAddress, req.Msg.ShippingMethodId, req.Msg.Note)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.CreateOrderResponse{
		OrderId: order.ID,
	}), nil
}

func (h *ProjectHandler) ListProjectOrders(ctx context.Context, req *connect.Request[v1.ListProjectOrdersRequest]) (*connect.Response[v1.ListProjectOrdersResponse], error) {
	orders, err := h.svc.ListProjectOrders(ctx, req.Msg.ProjectId)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var protoOrders []*v1.Order
	for _, o := range orders {
		protoOrders = append(protoOrders, toProtoOrder(o))
	}

	return connect.NewResponse(&v1.ListProjectOrdersResponse{
		Orders: protoOrders,
	}), nil
}

func (h *ProjectHandler) GetMyOrders(ctx context.Context, req *connect.Request[v1.GetMyOrdersRequest]) (*connect.Response[v1.GetMyOrdersResponse], error) {
	orders, err := h.svc.GetMyOrders(ctx)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
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

func (h *ProjectHandler) BatchUpdateStatus(ctx context.Context, req *connect.Request[v1.BatchUpdateStatusRequest]) (*connect.Response[v1.BatchUpdateStatusResponse], error) {
	updatedCount, updatedIds, err := h.svc.BatchUpdateStatus(ctx, req.Msg.ProjectId, req.Msg.SpecId, int(req.Msg.TargetStatus), req.Msg.Count)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.BatchUpdateStatusResponse{
		UpdatedCount:    updatedCount,
		UpdatedOrderIds: updatedIds,
	}), nil
}

func (h *ProjectHandler) ConfirmPayment(ctx context.Context, req *connect.Request[v1.ConfirmPaymentRequest]) (*connect.Response[v1.ConfirmPaymentResponse], error) {
	status := int(req.Msg.Status)
	if err := h.svc.ConfirmPayment(ctx, req.Msg.OrderId, status); err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.ConfirmPaymentResponse{
		OrderId: req.Msg.OrderId,
		Status:  req.Msg.Status,
	}), nil
}

func (h *ProjectHandler) GetMyProjectOrder(ctx context.Context, req *connect.Request[v1.GetMyProjectOrderRequest]) (*connect.Response[v1.GetMyProjectOrderResponse], error) {
	order, err := h.svc.GetMyProjectOrder(ctx, req.Msg.ProjectId)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var protoOrder *v1.Order
	if order != nil {
		protoOrder = toProtoOrder(order)
	}

	return connect.NewResponse(&v1.GetMyProjectOrderResponse{
		Order: protoOrder,
	}), nil
}

func (h *ProjectHandler) UpdateOrder(ctx context.Context, req *connect.Request[v1.UpdateOrderRequest]) (*connect.Response[v1.UpdateOrderResponse], error) {
	var items []*project.OrderItem
	for _, i := range req.Msg.Items {
		items = append(items, &project.OrderItem{
			ProductID: i.ProductId,
			SpecID:    i.SpecId,
			Quantity:  int(i.Quantity),
			Status:    int(i.Status),
		})
	}

	order, err := h.svc.UpdateOrder(ctx, req.Msg.OrderId, items, req.Msg.Note)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.UpdateOrderResponse{
		Order: toProtoOrder(order),
	}), nil
}

func (h *ProjectHandler) UpdatePaymentInfo(ctx context.Context, req *connect.Request[v1.UpdatePaymentInfoRequest]) (*connect.Response[v1.UpdatePaymentInfoResponse], error) {
	var paidAt *time.Time
	if req.Msg.PaidAt != nil {
		t := req.Msg.PaidAt.AsTime()
		paidAt = &t
	}
	order, err := h.svc.UpdatePaymentInfo(ctx, req.Msg.OrderId, req.Msg.Method, req.Msg.AccountLast5, req.Msg.ContactInfo, req.Msg.ShippingAddress, paidAt, req.Msg.Amount)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.UpdatePaymentInfoResponse{
		Order: toProtoOrder(order),
	}), nil
}

func (h *ProjectHandler) CancelOrder(ctx context.Context, req *connect.Request[v1.CancelOrderRequest]) (*connect.Response[v1.CancelOrderResponse], error) {
	if err := h.svc.CancelOrder(ctx, req.Msg.OrderId); err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.CancelOrderResponse{
		OrderId: req.Msg.OrderId,
		// Status: ...
	}), nil
}

func toProtoProject(p *project.Project) *v1.Project {
	if p == nil {
		return nil
	}

	var deadline *timestamppb.Timestamp
	if p.Deadline != nil {
		deadline = timestamppb.New(*p.Deadline)
	}

	return &v1.Project{
		Id:            p.ID,
		Title:         p.Title,
		Description:   p.Description,
		CoverImageUrl: p.CoverImage,
		Status:        v1.ProjectStatus(p.Status),
		CreatedAt:     timestamppb.New(p.CreatedAt),
		Deadline:      deadline,
		ExchangeRate:  p.ExchangeRate,
		RoundingConfig: &v1.RoundingConfig{
			Method: v1.RoundingMethod(p.Rounding.Method),
			Digit:  int32(p.Rounding.Digit),
		},
		Creator: toProtoUser(p.Creator),

		SourceCurrency:  p.SourceCurrency,
		ShippingConfigs: toProtoShippingConfigs(p.ShippingConfigs),
	}
}

func toProtoShippingConfigs(configs []*project.ShippingConfig) []*v1.ShippingConfig {
	var res []*v1.ShippingConfig
	for _, c := range configs {
		res = append(res, toProtoShippingConfig(c))
	}
	return res
}

func toProtoOrder(o *project.Order) *v1.Order {
	if o == nil {
		return nil
	}
	var items []*v1.OrderItem
	for _, i := range o.Items {
		items = append(items, &v1.OrderItem{
			Id:          i.ID,
			ProductId:   i.ProductID,
			SpecId:      i.SpecID,
			Quantity:    int32(i.Quantity),
			Status:      v1.OrderItemStatus(i.Status),
			ProductName: i.ProductName,
			SpecName:    i.SpecName,
			Price:       i.Price,
		})
	}

	var pi *v1.PaymentInfo
	if o.PaymentInfo != nil {
		var paidAt *timestamppb.Timestamp
		if o.PaymentInfo.PaidAt != nil {
			paidAt = timestamppb.New(*o.PaymentInfo.PaidAt)
		}
		pi = &v1.PaymentInfo{
			Method:       o.PaymentInfo.Method,
			AccountLast5: o.PaymentInfo.AccountLast5,
			PaidAt:       paidAt,
			Amount:       o.PaymentInfo.Amount,
		}
	}

	return &v1.Order{
		Id:               o.ID,
		ProjectId:        o.ProjectID,
		UserId:           o.UserID,
		TotalAmount:      o.TotalAmount,
		PaymentStatus:    v1.PaymentStatus(o.PaymentStatus),
		ContactInfo:      o.ContactInfo,
		ShippingAddress:  o.ShippingAddress,
		PaymentInfo:      pi,
		Items:            items,
		ShippingMethodId: o.ShippingMethodID,
		ShippingFee:      o.ShippingFee,
		Note:             o.Note,
	}
}

func fromProtoShippingConfig(c *v1.ShippingConfig) *project.ShippingConfig {
	if c == nil {
		return nil
	}
	return &project.ShippingConfig{
		ID:    c.Id,
		Name:  c.Name,
		Type:  project.ShippingType(c.Type),
		Price: c.Price,
	}
}

func toProtoShippingConfig(c *project.ShippingConfig) *v1.ShippingConfig {
	if c == nil {
		return nil
	}
	return &v1.ShippingConfig{
		Id:    c.ID,
		Name:  c.Name,
		Type:  v1.ShippingType(c.Type),
		Price: c.Price,
	}
}

func toProtoProducts(products []*project.Product) []*v1.Product {
	var res []*v1.Product
	for _, p := range products {
		res = append(res, toProtoProduct(p))
	}
	return res
}

func toProtoProduct(p *project.Product) *v1.Product {
	if p == nil {
		return nil
	}
	var specs []*v1.ProductSpec
	for _, s := range p.Specs {
		specs = append(specs, &v1.ProductSpec{Id: s.ID, Name: s.Name})
	}
	var rc *v1.RoundingConfig
	if p.Rounding != nil {
		rc = &v1.RoundingConfig{
			Method: v1.RoundingMethod(p.Rounding.Method),
			Digit:  int32(p.Rounding.Digit),
		}
	}
	return &v1.Product{
		Id:             p.ID,
		ProjectId:      p.ProjectID,
		Name:           p.Name,
		Description:    p.Description,
		ImageUrl:       p.ImageURL,
		PriceOriginal:  p.PriceOriginal,
		ExchangeRate:   p.ExchangeRate,
		RoundingConfig: rc,
		PriceFinal:     p.PriceFinal,
		MaxQuantity:    p.MaxQuantity,
		Specs:          specs,
	}
}
func (h *ProjectHandler) CreateCategory(ctx context.Context, req *connect.Request[v1.CreateCategoryRequest]) (*connect.Response[v1.CreateCategoryResponse], error) {
	c, err := h.svc.CreateCategory(ctx, req.Msg.Name, req.Msg.SpecNames)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.CreateCategoryResponse{
		Category: toProtoCategory(c),
	}), nil
}

func (h *ProjectHandler) ListCategories(ctx context.Context, req *connect.Request[v1.ListCategoriesRequest]) (*connect.Response[v1.ListCategoriesResponse], error) {
	categories, err := h.svc.ListCategories(ctx)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var protoCategories []*v1.Category
	for _, c := range categories {
		protoCategories = append(protoCategories, toProtoCategory(c))
	}

	return connect.NewResponse(&v1.ListCategoriesResponse{
		Categories: protoCategories,
	}), nil
}

func toProtoCategory(c *project.Category) *v1.Category {
	if c == nil {
		return nil
	}
	return &v1.Category{
		Id:        c.ID,
		Name:      c.Name,
		SpecNames: c.SpecNames,
	}
}

// PriceTemplate Handlers

func (h *ProjectHandler) CreatePriceTemplate(ctx context.Context, req *connect.Request[v1.CreatePriceTemplateRequest]) (*connect.Response[v1.CreatePriceTemplateResponse], error) {
	var rounding *project.RoundingConfig
	if req.Msg.RoundingConfig != nil {
		rounding = &project.RoundingConfig{
			Method: int(req.Msg.RoundingConfig.Method),
			Digit:  int(req.Msg.RoundingConfig.Digit),
		}
	}

	pt, err := h.svc.CreatePriceTemplate(ctx, req.Msg.Name, req.Msg.SourceCurrency, req.Msg.ExchangeRate, rounding)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.CreatePriceTemplateResponse{
		Template: toProtoPriceTemplate(pt),
	}), nil
}

func (h *ProjectHandler) ListPriceTemplates(ctx context.Context, req *connect.Request[v1.ListPriceTemplatesRequest]) (*connect.Response[v1.ListPriceTemplatesResponse], error) {
	templates, err := h.svc.ListPriceTemplates(ctx)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var protoTemplates []*v1.PriceTemplate
	for _, pt := range templates {
		protoTemplates = append(protoTemplates, toProtoPriceTemplate(pt))
	}

	return connect.NewResponse(&v1.ListPriceTemplatesResponse{
		Templates: protoTemplates,
	}), nil
}

func (h *ProjectHandler) GetPriceTemplate(ctx context.Context, req *connect.Request[v1.GetPriceTemplateRequest]) (*connect.Response[v1.GetPriceTemplateResponse], error) {
	pt, err := h.svc.GetPriceTemplate(ctx, req.Msg.TemplateId)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.GetPriceTemplateResponse{
		Template: toProtoPriceTemplate(pt),
	}), nil
}

func (h *ProjectHandler) UpdatePriceTemplate(ctx context.Context, req *connect.Request[v1.UpdatePriceTemplateRequest]) (*connect.Response[v1.UpdatePriceTemplateResponse], error) {
	var rounding *project.RoundingConfig
	if req.Msg.RoundingConfig != nil {
		rounding = &project.RoundingConfig{
			Method: int(req.Msg.RoundingConfig.Method),
			Digit:  int(req.Msg.RoundingConfig.Digit),
		}
	}

	pt, err := h.svc.UpdatePriceTemplate(ctx, req.Msg.TemplateId, req.Msg.Name, req.Msg.SourceCurrency, req.Msg.ExchangeRate, rounding)
	if err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.UpdatePriceTemplateResponse{
		Template: toProtoPriceTemplate(pt),
	}), nil
}

func (h *ProjectHandler) DeletePriceTemplate(ctx context.Context, req *connect.Request[v1.DeletePriceTemplateRequest]) (*connect.Response[v1.DeletePriceTemplateResponse], error) {
	if err := h.svc.DeletePriceTemplate(ctx, req.Msg.TemplateId); err != nil {
		if errors.Is(err, service.ErrPermissionDenied) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.DeletePriceTemplateResponse{
		TemplateId: req.Msg.TemplateId,
	}), nil
}

func toProtoPriceTemplate(pt *project.PriceTemplate) *v1.PriceTemplate {
	if pt == nil {
		return nil
	}
	var rc *v1.RoundingConfig
	if pt.Rounding != nil {
		rc = &v1.RoundingConfig{
			Method: v1.RoundingMethod(pt.Rounding.Method),
			Digit:  int32(pt.Rounding.Digit),
		}
	}
	return &v1.PriceTemplate{
		Id:             pt.ID,
		Name:           pt.Name,
		SourceCurrency: pt.SourceCurrency,
		ExchangeRate:   pt.ExchangeRate,
		RoundingConfig: rc,
	}
}
