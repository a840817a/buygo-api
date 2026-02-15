package handler

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/buygo/buygo-api/api/v1"
	"github.com/buygo/buygo-api/api/v1/buygov1connect"
	"github.com/buygo/buygo-api/internal/domain/event"
	"github.com/buygo/buygo-api/internal/service"
)

type EventHandler struct {
	svc *service.EventService
}

func NewEventHandler(svc *service.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

// Ensure implementation
var _ buygov1connect.EventServiceHandler = (*EventHandler)(nil)

func (h *EventHandler) CreateEvent(ctx context.Context, req *connect.Request[v1.CreateEventRequest]) (*connect.Response[v1.CreateEventResponse], error) {
	start := req.Msg.StartTime.AsTime()
	end := req.Msg.EndTime.AsTime()

	var registrationDeadline *time.Time
	if req.Msg.RegistrationDeadline != nil {
		t := req.Msg.RegistrationDeadline.AsTime()
		registrationDeadline = &t
	}

	var discounts []*event.DiscountRule
	for _, d := range req.Msg.Discounts {
		discounts = append(discounts, &event.DiscountRule{
			MinQuantity:      int32(d.MinQuantity),
			MinDistinctItems: int32(d.MinDistinctItems),
			DiscountAmount:   int64(d.DiscountAmount),
		})
	}

	var items []*event.EventItem
	for _, i := range req.Msg.Items {
		items = append(items, &event.EventItem{
			ID:              i.Id,
			Name:            i.Name,
			Price:           i.Price,
			MinParticipants: i.MinParticipants,
			MaxParticipants: i.MaxParticipants,
			StartTime:       toTime(i.StartTime),
			EndTime:         toTime(i.EndTime),
			AllowMultiple:   i.AllowMultiple,
		})
	}

	e, err := h.svc.CreateEvent(ctx,
		req.Msg.Title,
		req.Msg.Description,
		req.Msg.Location,
		req.Msg.CoverImageUrl,
		start, end,
		registrationDeadline,
		req.Msg.PaymentMethods,
		req.Msg.AllowModification,
		req.Msg.ManagerIds,
		items,
		discounts,
	)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.CreateEventResponse{
		Event: toProtoEvent(e),
	}), nil
}

func (h *EventHandler) ListEvents(ctx context.Context, req *connect.Request[v1.ListEventsRequest]) (*connect.Response[v1.ListEventsResponse], error) {
	events, err := h.svc.ListEvents(ctx, int(req.Msg.PageSize), 0)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var protoEvents []*v1.Event
	for _, e := range events {
		protoEvents = append(protoEvents, toProtoEvent(e))
	}

	return connect.NewResponse(&v1.ListEventsResponse{
		Events: protoEvents,
	}), nil
}

func (h *EventHandler) ListManagerEvents(ctx context.Context, req *connect.Request[v1.ListManagerEventsRequest]) (*connect.Response[v1.ListManagerEventsResponse], error) {
	events, err := h.svc.ListManagerEvents(ctx, int(req.Msg.PageSize), 0)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var protoEvents []*v1.Event
	for _, e := range events {
		protoEvents = append(protoEvents, toProtoEvent(e))
	}

	return connect.NewResponse(&v1.ListManagerEventsResponse{
		Events: protoEvents,
	}), nil
}

func (h *EventHandler) GetEvent(ctx context.Context, req *connect.Request[v1.GetEventRequest]) (*connect.Response[v1.GetEventResponse], error) {
	e, err := h.svc.GetEvent(ctx, req.Msg.EventId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&v1.GetEventResponse{
		Event: toProtoEvent(e),
	}), nil
}

func (h *EventHandler) RegisterEvent(ctx context.Context, req *connect.Request[v1.RegisterEventRequest]) (*connect.Response[v1.RegisterEventResponse], error) {
	var items []*event.RegistrationItem
	for _, i := range req.Msg.Items {
		items = append(items, &event.RegistrationItem{
			EventItemID: i.EventItemId,
			Quantity:    int(i.Quantity),
		})
	}

	reg, err := h.svc.RegisterEvent(ctx, req.Msg.EventId, items, req.Msg.ContactInfo, req.Msg.Notes)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.RegisterEventResponse{
		RegistrationId: reg.ID,
		Status:         v1.RegistrationStatus(reg.Status),
	}), nil
}

func (h *EventHandler) UpdateRegistration(ctx context.Context, req *connect.Request[v1.UpdateRegistrationRequest]) (*connect.Response[v1.UpdateRegistrationResponse], error) {
	var items []*event.RegistrationItem
	for _, i := range req.Msg.Items {
		items = append(items, &event.RegistrationItem{
			EventItemID: i.EventItemId,
			Quantity:    int(i.Quantity),
		})
	}

	reg, err := h.svc.UpdateRegistration(ctx, req.Msg.RegistrationId, items, req.Msg.ContactInfo, req.Msg.Notes)
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.UpdateRegistrationResponse{
		RegistrationId: reg.ID,
		Status:         v1.RegistrationStatus(reg.Status),
	}), nil
}

func (h *EventHandler) UpdateRegistrationStatus(ctx context.Context, req *connect.Request[v1.UpdateRegistrationStatusRequest]) (*connect.Response[v1.UpdateRegistrationStatusResponse], error) {
	reg, err := h.svc.UpdateRegistrationStatus(ctx, req.Msg.RegistrationId, event.RegistrationStatus(req.Msg.Status), event.PaymentStatus(req.Msg.PaymentStatus))
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.UpdateRegistrationStatusResponse{
		RegistrationId: reg.ID,
		Status:         v1.RegistrationStatus(reg.Status),
		PaymentStatus:  v1.PaymentStatus(reg.PaymentStatus),
	}), nil
}

func (h *EventHandler) CancelRegistration(ctx context.Context, req *connect.Request[v1.CancelRegistrationRequest]) (*connect.Response[v1.CancelRegistrationResponse], error) {
	if err := h.svc.CancelRegistration(ctx, req.Msg.RegistrationId); err != nil {
		return nil, mapError(err)
	}
	return connect.NewResponse(&v1.CancelRegistrationResponse{
		RegistrationId: req.Msg.RegistrationId,
		Status:         v1.RegistrationStatus_REGISTRATION_STATUS_CANCELLED,
	}), nil
}

func (h *EventHandler) GetMyRegistrations(ctx context.Context, req *connect.Request[v1.GetMyRegistrationsRequest]) (*connect.Response[v1.GetMyRegistrationsResponse], error) {
	regs, err := h.svc.GetMyRegistrations(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	var protoRegs []*v1.Registration
	for _, r := range regs {
		protoRegs = append(protoRegs, toProtoRegistration(r))
	}

	return connect.NewResponse(&v1.GetMyRegistrationsResponse{
		Registrations: protoRegs,
	}), nil
}

func (h *EventHandler) ListEventRegistrations(ctx context.Context, req *connect.Request[v1.ListEventRegistrationsRequest]) (*connect.Response[v1.ListEventRegistrationsResponse], error) {
	regs, err := h.svc.ListEventRegistrations(ctx, req.Msg.EventId)
	if err != nil {
		return nil, mapError(err)
	}

	var protoRegs []*v1.Registration
	for _, r := range regs {
		protoRegs = append(protoRegs, toProtoRegistration(r))
	}

	return connect.NewResponse(&v1.ListEventRegistrationsResponse{
		Registrations: protoRegs,
	}), nil
}

func (h *EventHandler) UpdateEvent(ctx context.Context, req *connect.Request[v1.UpdateEventRequest]) (*connect.Response[v1.UpdateEventResponse], error) {
	// Convert items
	var items []*event.EventItem
	for _, i := range req.Msg.Items {
		items = append(items, &event.EventItem{
			ID:              i.Id,
			Name:            i.Name,
			Price:           i.Price,
			MinParticipants: i.MinParticipants,
			MaxParticipants: i.MaxParticipants,
			StartTime:       toTime(i.StartTime),
			EndTime:         toTime(i.EndTime),
			AllowMultiple:   i.AllowMultiple,
		})
	}

	var discounts []*event.DiscountRule
	for _, d := range req.Msg.Discounts {
		discounts = append(discounts, &event.DiscountRule{
			MinQuantity:      int32(d.MinQuantity),
			MinDistinctItems: int32(d.MinDistinctItems),
			DiscountAmount:   int64(d.DiscountAmount),
		})
	}

	e, err := h.svc.UpdateEvent(ctx,
		req.Msg.EventId,
		req.Msg.Title,
		req.Msg.Description,
		req.Msg.Location,
		req.Msg.CoverImageUrl,
		req.Msg.StartTime.AsTime(),
		req.Msg.EndTime.AsTime(),
		req.Msg.AllowModification,
		items,
		req.Msg.ManagerIds,
		discounts,
	)

	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.UpdateEventResponse{
		Event: toProtoEvent(e),
	}), nil
}

func (h *EventHandler) UpdateEventStatus(ctx context.Context, req *connect.Request[v1.UpdateEventStatusRequest]) (*connect.Response[v1.UpdateEventStatusResponse], error) {
	e, err := h.svc.UpdateEventStatus(ctx, req.Msg.EventId, event.EventStatus(req.Msg.Status))
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.UpdateEventStatusResponse{
		Event: toProtoEvent(e),
	}), nil
}

func toProtoEvent(e *event.Event) *v1.Event {
	if e == nil {
		return nil
	}

	var managers []*v1.User
	for _, m := range e.Managers {
		managers = append(managers, toProtoUser(m))
	}

	var items []*v1.EventItem
	for _, i := range e.Items {
		items = append(items, toProtoEventItem(i))
	}

	return &v1.Event{
		Id:                   e.ID,
		Title:                e.Title,
		Description:          e.Description,
		CoverImageUrl:        e.CoverImage,
		Status:               v1.EventStatus(e.Status),
		StartTime:            timestamppb.New(e.StartTime),
		EndTime:              timestamppb.New(e.EndTime),
		RegistrationDeadline: timestamppb.New(e.RegistrationDeadline),
		Location:             e.Location,
		Creator:              toProtoUser(e.Creator),
		Managers:             managers,
		PaymentMethods:       e.PaymentMethods,
		Items:                items,
		AllowModification:    e.AllowException,
		Discounts:            toProtoDiscounts(e.Discounts),
	}
}

func toProtoDiscounts(rules []*event.DiscountRule) []*v1.DiscountRule {
	var res []*v1.DiscountRule
	for _, r := range rules {
		res = append(res, &v1.DiscountRule{
			MinQuantity:      int32(r.MinQuantity),
			MinDistinctItems: int32(r.MinDistinctItems),
			DiscountAmount:   int64(r.DiscountAmount),
		})
	}
	return res
}

func toTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func toProtoEventItem(i *event.EventItem) *v1.EventItem {
	if i == nil {
		return nil
	}
	var start, end *timestamppb.Timestamp
	if i.StartTime != nil {
		start = timestamppb.New(*i.StartTime)
	}
	if i.EndTime != nil {
		end = timestamppb.New(*i.EndTime)
	}

	return &v1.EventItem{
		Id:              i.ID,
		Name:            i.Name,
		Price:           i.Price,
		MinParticipants: i.MinParticipants,
		MaxParticipants: i.MaxParticipants,
		StartTime:       start,
		EndTime:         end,
		AllowMultiple:   i.AllowMultiple,
	}
}

func toProtoRegistration(r *event.Registration) *v1.Registration {
	if r == nil {
		return nil
	}

	// Flatten items for proto if needed or keep structure
	// Proto expects RegisterItem which matches logic
	var items []*v1.RegisterItem
	for _, i := range r.SelectedItems {
		items = append(items, &v1.RegisterItem{
			EventItemId: i.EventItemID,
			Quantity:    int32(i.Quantity),
		})
	}

	return &v1.Registration{
		Id:              r.ID,
		EventId:         r.EventID,
		UserId:          r.UserID,
		Status:          v1.RegistrationStatus(r.Status),
		PaymentStatus:   v1.PaymentStatus(r.PaymentStatus),
		ContactInfo:     r.ContactInfo,
		Notes:           r.Notes,
		TotalAmount:     r.TotalAmount,
		DiscountApplied: r.DiscountApplied,
		SelectedItems:   items,
		User:            toProtoUser(r.User),
	}
}
