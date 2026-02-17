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
	start := time.Now().Add(-1 * time.Hour) // 已開始
	end := time.Now().Add(24 * time.Hour)

	createResp, _ := h.CreateEvent(creatorCtx, connect.NewRequest(&v1.CreateEventRequest{
		Title:     "Active Event",
		StartTime: timestamppb.New(start),
		EndTime:   timestamppb.New(end),
	}))
	eventID := createResp.Msg.Event.Id

	// 啟用活動
	_, _ = h.UpdateEventStatus(creatorCtx, connect.NewRequest(&v1.UpdateEventStatusRequest{
		EventId: eventID,
		Status:  v1.EventStatus_EVENT_STATUS_ACTIVE,
	}))

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
}

func TestEventHandler_ListManagerEvents(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	managerCtx := domainAuth.NewContext(context.Background(), "manager-1", int(user.UserRoleCreator))
	_, _ = h.CreateEvent(managerCtx, connect.NewRequest(&v1.CreateEventRequest{
		Title:      "Event 1",
		ManagerIds: []string{"manager-1"},
	}))

	resp, err := h.ListManagerEvents(managerCtx, connect.NewRequest(&v1.ListManagerEventsRequest{
		PageSize: 10,
	}))
	if err != nil {
		t.Fatalf("ListManagerEvents error: %v", err)
	}

	if len(resp.Msg.Events) != 1 {
		t.Errorf("got %d events, want 1", len(resp.Msg.Events))
	}
}
