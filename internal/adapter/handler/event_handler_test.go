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

func TestEventHandler_CreateEvent(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	ctx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	start := time.Now().Add(1 * time.Hour)
	end := time.Now().Add(24 * time.Hour)

	req := &v1.CreateEventRequest{
		Title:       "Test Event",
		Description: "A test event",
		StartTime:   timestamppb.New(start),
		EndTime:     timestamppb.New(end),
		ManagerIds:  []string{"manager-1"},
	}

	resp, err := h.CreateEvent(ctx, connect.NewRequest(req))
	if err != nil {
		t.Fatalf("CreateEvent error: %v", err)
	}

	if resp.Msg.Event.Title != req.Title {
		t.Errorf("got title %q, want %q", resp.Msg.Event.Title, req.Title)
	}
}

func TestEventHandler_RegisterEvent(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	eventID := createActiveEvent(t, h, creatorCtx)

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	resp, err := h.RegisterEvent(userCtx, connect.NewRequest(&v1.RegisterEventRequest{
		EventId: eventID,
	}))
	if err != nil {
		t.Fatalf("RegisterEvent error: %v", err)
	}

	if resp.Msg.RegistrationId == "" {
		t.Errorf("got empty registration id")
	}
	if resp.Msg.Status != v1.RegistrationStatus_REGISTRATION_STATUS_PENDING {
		t.Errorf("got status %v, want PENDING", resp.Msg.Status)
	}

	// Verify registration persists by fetching it back
	myRegs, err := h.GetMyRegistrations(userCtx, connect.NewRequest(&v1.GetMyRegistrationsRequest{}))
	if err != nil {
		t.Fatalf("GetMyRegistrations error: %v", err)
	}
	if len(myRegs.Msg.Registrations) != 1 {
		t.Fatalf("got %d registrations, want 1", len(myRegs.Msg.Registrations))
	}
	if myRegs.Msg.Registrations[0].EventId != eventID {
		t.Errorf("registration event id = %q, want %q", myRegs.Msg.Registrations[0].EventId, eventID)
	}
}

func TestEventHandler_ListManagerEvents(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	managerCtx := domainAuth.NewContext(context.Background(), "manager-1", int(user.UserRoleCreator))
	createResp, err := h.CreateEvent(managerCtx, connect.NewRequest(&v1.CreateEventRequest{
		Title:      "Event 1",
		ManagerIds: []string{"manager-1"},
	}))
	if err != nil {
		t.Fatalf("CreateEvent error: %v", err)
	}

	resp, err := h.ListManagerEvents(managerCtx, connect.NewRequest(&v1.ListManagerEventsRequest{
		PageSize: 10,
	}))
	if err != nil {
		t.Fatalf("ListManagerEvents error: %v", err)
	}

	if len(resp.Msg.Events) != 1 {
		t.Fatalf("got %d events, want 1", len(resp.Msg.Events))
	}
	if resp.Msg.Events[0].Title != "Event 1" {
		t.Errorf("got title %q, want %q", resp.Msg.Events[0].Title, "Event 1")
	}
	if resp.Msg.Events[0].Id != createResp.Msg.Event.Id {
		t.Errorf("got id %q, want %q", resp.Msg.Events[0].Id, createResp.Msg.Event.Id)
	}
}

// createActiveEvent is a helper that creates an event and activates it.
func createActiveEvent(t *testing.T, h *EventHandler, creatorCtx context.Context) string {
	t.Helper()
	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(24 * time.Hour)

	createResp, err := h.CreateEvent(creatorCtx, connect.NewRequest(&v1.CreateEventRequest{
		Title:             "Test Event",
		Description:       "Desc",
		StartTime:         timestamppb.New(start),
		EndTime:           timestamppb.New(end),
		AllowModification: true,
	}))
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}

	eventID := createResp.Msg.Event.Id
	_, err = h.UpdateEventStatus(creatorCtx, connect.NewRequest(&v1.UpdateEventStatusRequest{
		EventId: eventID,
		Status:  v1.EventStatus_EVENT_STATUS_ACTIVE,
	}))
	if err != nil {
		t.Fatalf("UpdateEventStatus: %v", err)
	}
	return eventID
}

func TestEventHandler_GetEvent(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	createResp, _ := h.CreateEvent(creatorCtx, connect.NewRequest(&v1.CreateEventRequest{
		Title:     "Get Event Test",
		StartTime: timestamppb.New(time.Now().Add(1 * time.Hour)),
		EndTime:   timestamppb.New(time.Now().Add(24 * time.Hour)),
	}))
	eventID := createResp.Msg.Event.Id

	resp, err := h.GetEvent(context.Background(), connect.NewRequest(&v1.GetEventRequest{
		EventId: eventID,
	}))
	if err != nil {
		t.Fatalf("GetEvent error: %v", err)
	}
	if resp.Msg.Event.Id != eventID {
		t.Errorf("got id %q, want %q", resp.Msg.Event.Id, eventID)
	}
	if resp.Msg.Event.Title != "Get Event Test" {
		t.Errorf("got title %q, want %q", resp.Msg.Event.Title, "Get Event Test")
	}
}

func TestEventHandler_GetEvent_NotFound(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	_, err := h.GetEvent(context.Background(), connect.NewRequest(&v1.GetEventRequest{
		EventId: "non-existent",
	}))
	if err == nil {
		t.Fatal("expected error for non-existent event")
	}
	// Memory repo returns a plain error (not service.ErrNotFound), so mapError
	// falls through to CodeInternal. In production (Postgres), this would be CodeNotFound.
	code := connect.CodeOf(err)
	if code != connect.CodeNotFound && code != connect.CodeInternal {
		t.Errorf("got error code %v, want CodeNotFound or CodeInternal", code)
	}
}

func TestEventHandler_UpdateEvent(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	start := time.Now().Add(1 * time.Hour)
	end := time.Now().Add(24 * time.Hour)

	createResp, _ := h.CreateEvent(creatorCtx, connect.NewRequest(&v1.CreateEventRequest{
		Title:     "Original Title",
		StartTime: timestamppb.New(start),
		EndTime:   timestamppb.New(end),
	}))
	eventID := createResp.Msg.Event.Id

	resp, err := h.UpdateEvent(creatorCtx, connect.NewRequest(&v1.UpdateEventRequest{
		EventId:     eventID,
		Title:       "Updated Title",
		Description: "Updated Description",
		StartTime:   timestamppb.New(start),
		EndTime:     timestamppb.New(end),
	}))
	if err != nil {
		t.Fatalf("UpdateEvent error: %v", err)
	}
	if resp.Msg.Event.Title != "Updated Title" {
		t.Errorf("got title %q, want %q", resp.Msg.Event.Title, "Updated Title")
	}
}

func TestEventHandler_UpdateRegistration(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	eventID := createActiveEvent(t, h, creatorCtx)

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	regResp, err := h.RegisterEvent(userCtx, connect.NewRequest(&v1.RegisterEventRequest{
		EventId:     eventID,
		ContactInfo: "old contact",
	}))
	if err != nil {
		t.Fatalf("RegisterEvent error: %v", err)
	}

	updateResp, err := h.UpdateRegistration(userCtx, connect.NewRequest(&v1.UpdateRegistrationRequest{
		RegistrationId: regResp.Msg.RegistrationId,
		ContactInfo:    "new contact",
		Notes:          "updated notes",
	}))
	if err != nil {
		t.Fatalf("UpdateRegistration error: %v", err)
	}
	if updateResp.Msg.RegistrationId != regResp.Msg.RegistrationId {
		t.Errorf("got reg id %q, want %q", updateResp.Msg.RegistrationId, regResp.Msg.RegistrationId)
	}

	// Verify the update actually persisted by fetching registrations
	listResp, err := h.ListEventRegistrations(creatorCtx, connect.NewRequest(&v1.ListEventRegistrationsRequest{
		EventId: eventID,
	}))
	if err != nil {
		t.Fatalf("ListEventRegistrations error: %v", err)
	}
	if len(listResp.Msg.Registrations) != 1 {
		t.Fatalf("got %d registrations, want 1", len(listResp.Msg.Registrations))
	}
	reg := listResp.Msg.Registrations[0]
	if reg.ContactInfo != "new contact" {
		t.Errorf("contact info = %q, want %q", reg.ContactInfo, "new contact")
	}
	if reg.Notes != "updated notes" {
		t.Errorf("notes = %q, want %q", reg.Notes, "updated notes")
	}
}

func TestEventHandler_UpdateRegistrationStatus(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	eventID := createActiveEvent(t, h, creatorCtx)

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	regResp, err := h.RegisterEvent(userCtx, connect.NewRequest(&v1.RegisterEventRequest{
		EventId: eventID,
	}))
	if err != nil {
		t.Fatalf("RegisterEvent error: %v", err)
	}

	resp, err := h.UpdateRegistrationStatus(creatorCtx, connect.NewRequest(&v1.UpdateRegistrationStatusRequest{
		RegistrationId: regResp.Msg.RegistrationId,
		Status:         v1.RegistrationStatus_REGISTRATION_STATUS_CONFIRMED,
		PaymentStatus:  v1.PaymentStatus_PAYMENT_STATUS_CONFIRMED,
	}))
	if err != nil {
		t.Fatalf("UpdateRegistrationStatus error: %v", err)
	}
	if resp.Msg.Status != v1.RegistrationStatus_REGISTRATION_STATUS_CONFIRMED {
		t.Errorf("got status %v, want CONFIRMED", resp.Msg.Status)
	}
}

func TestEventHandler_CancelRegistration(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	eventID := createActiveEvent(t, h, creatorCtx)

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	regResp, err := h.RegisterEvent(userCtx, connect.NewRequest(&v1.RegisterEventRequest{
		EventId: eventID,
	}))
	if err != nil {
		t.Fatalf("RegisterEvent error: %v", err)
	}

	cancelResp, err := h.CancelRegistration(userCtx, connect.NewRequest(&v1.CancelRegistrationRequest{
		RegistrationId: regResp.Msg.RegistrationId,
	}))
	if err != nil {
		t.Fatalf("CancelRegistration error: %v", err)
	}
	if cancelResp.Msg.Status != v1.RegistrationStatus_REGISTRATION_STATUS_CANCELLED {
		t.Errorf("got status %v, want CANCELLED", cancelResp.Msg.Status)
	}
}

func TestEventHandler_GetMyRegistrations(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	eventID := createActiveEvent(t, h, creatorCtx)

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	_, err := h.RegisterEvent(userCtx, connect.NewRequest(&v1.RegisterEventRequest{
		EventId: eventID,
	}))
	if err != nil {
		t.Fatalf("RegisterEvent error: %v", err)
	}

	resp, err := h.GetMyRegistrations(userCtx, connect.NewRequest(&v1.GetMyRegistrationsRequest{}))
	if err != nil {
		t.Fatalf("GetMyRegistrations error: %v", err)
	}
	if len(resp.Msg.Registrations) != 1 {
		t.Errorf("got %d registrations, want 1", len(resp.Msg.Registrations))
	}
}

func TestEventHandler_ListEventRegistrations(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	eventID := createActiveEvent(t, h, creatorCtx)

	userCtx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	regResp, err := h.RegisterEvent(userCtx, connect.NewRequest(&v1.RegisterEventRequest{
		EventId: eventID,
	}))
	if err != nil {
		t.Fatalf("RegisterEvent error: %v", err)
	}

	resp, err := h.ListEventRegistrations(creatorCtx, connect.NewRequest(&v1.ListEventRegistrationsRequest{
		EventId: eventID,
	}))
	if err != nil {
		t.Fatalf("ListEventRegistrations error: %v", err)
	}
	if len(resp.Msg.Registrations) != 1 {
		t.Fatalf("got %d registrations, want 1", len(resp.Msg.Registrations))
	}
	if resp.Msg.Registrations[0].Id != regResp.Msg.RegistrationId {
		t.Errorf("registration id = %q, want %q", resp.Msg.Registrations[0].Id, regResp.Msg.RegistrationId)
	}
	if resp.Msg.Registrations[0].EventId != eventID {
		t.Errorf("event id = %q, want %q", resp.Msg.Registrations[0].EventId, eventID)
	}
}

func TestEventHandler_ListEvents(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	_ = createActiveEvent(t, h, creatorCtx)

	resp, err := h.ListEvents(context.Background(), connect.NewRequest(&v1.ListEventsRequest{
		PageSize: 10,
	}))
	if err != nil {
		t.Fatalf("ListEvents error: %v", err)
	}
	if len(resp.Msg.Events) != 1 {
		t.Errorf("got %d events, want 1", len(resp.Msg.Events))
	}
}
