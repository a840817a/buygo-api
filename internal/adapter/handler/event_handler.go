package handler

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/hatsubosi/buygo-api/api/v1"
	"github.com/hatsubosi/buygo-api/api/v1/buygov1connect"
	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/hatsubosi/buygo-api/internal/service"
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
	limit := normalizePageSize(int(req.Msg.PageSize))
	offset, err := decodePageToken(req.Msg.PageToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid page_token"))
	}

	events, err := h.svc.ListEvents(ctx, limit+1, offset)
	if err != nil {
		return nil, mapError(err)
	}

	nextPageToken := ""
	if len(events) > limit {
		events = events[:limit]
		nextPageToken = encodePageToken(offset + limit)
	}

	var protoEvents []*v1.Event
	for _, e := range events {
		protoEvents = append(protoEvents, toProtoEvent(e))
	}

	return connect.NewResponse(&v1.ListEventsResponse{
		Events:        protoEvents,
		NextPageToken: nextPageToken,
	}), nil
}

func (h *EventHandler) ListManagerEvents(ctx context.Context, req *connect.Request[v1.ListManagerEventsRequest]) (*connect.Response[v1.ListManagerEventsResponse], error) {
	limit := normalizePageSize(int(req.Msg.PageSize))
	offset, err := decodePageToken(req.Msg.PageToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid page_token"))
	}

	events, err := h.svc.ListManagerEvents(ctx, limit+1, offset)
	if err != nil {
		return nil, mapError(err)
	}

	nextPageToken := ""
	if len(events) > limit {
		events = events[:limit]
		nextPageToken = encodePageToken(offset + limit)
	}

	var protoEvents []*v1.Event
	for _, e := range events {
		protoEvents = append(protoEvents, toProtoEvent(e))
	}

	return connect.NewResponse(&v1.ListManagerEventsResponse{
		Events:        protoEvents,
		NextPageToken: nextPageToken,
	}), nil
}

func (h *EventHandler) GetEvent(ctx context.Context, req *connect.Request[v1.GetEventRequest]) (*connect.Response[v1.GetEventResponse], error) {
	e, err := h.svc.GetEvent(ctx, req.Msg.EventId)
	if err != nil {
		return nil, mapError(err)
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
		Status:         toProtoRegistrationStatus(reg.Status),
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
		Status:         toProtoRegistrationStatus(reg.Status),
	}), nil
}

func (h *EventHandler) UpdateRegistrationStatus(ctx context.Context, req *connect.Request[v1.UpdateRegistrationStatusRequest]) (*connect.Response[v1.UpdateRegistrationStatusResponse], error) {
	reg, err := h.svc.UpdateRegistrationStatus(ctx, req.Msg.RegistrationId, event.RegistrationStatus(req.Msg.Status), event.PaymentStatus(req.Msg.PaymentStatus))
	if err != nil {
		return nil, mapError(err)
	}

	return connect.NewResponse(&v1.UpdateRegistrationStatusResponse{
		RegistrationId: reg.ID,
		Status:         toProtoRegistrationStatus(reg.Status),
		PaymentStatus:  toProtoEventPaymentStatus(reg.PaymentStatus),
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
		Status:               toProtoEventStatus(e.Status),
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
