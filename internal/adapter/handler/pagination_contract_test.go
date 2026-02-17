package handler

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/hatsubosi/buygo-api/api/v1"
	"github.com/hatsubosi/buygo-api/internal/adapter/repository/memory"
	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/service"
)

func TestGroupBuyHandler_ListGroupBuys_PaginationContract(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	now := time.Now()
	_ = repo.Create(context.Background(), &groupbuy.GroupBuy{
		ID:        "gb-1",
		Title:     "GB-1",
		Status:    groupbuy.GroupBuyStatusActive,
		CreatedAt: now.Add(-1 * time.Hour),
	})
	_ = repo.Create(context.Background(), &groupbuy.GroupBuy{
		ID:        "gb-2",
		Title:     "GB-2",
		Status:    groupbuy.GroupBuyStatusActive,
		CreatedAt: now.Add(-2 * time.Hour),
	})
	_ = repo.Create(context.Background(), &groupbuy.GroupBuy{
		ID:        "gb-3",
		Title:     "GB-3",
		Status:    groupbuy.GroupBuyStatusActive,
		CreatedAt: now.Add(-3 * time.Hour),
	})

	firstResp, err := h.ListGroupBuys(context.Background(), connect.NewRequest(&v1.ListGroupBuysRequest{
		PageSize: 2,
	}))
	if err != nil {
		t.Fatalf("ListGroupBuys first page error: %v", err)
	}
	if len(firstResp.Msg.GroupBuys) != 2 {
		t.Fatalf("first page length = %d, want 2", len(firstResp.Msg.GroupBuys))
	}
	if firstResp.Msg.NextPageToken != "2" {
		t.Fatalf("next page token = %q, want %q", firstResp.Msg.NextPageToken, "2")
	}

	secondResp, err := h.ListGroupBuys(context.Background(), connect.NewRequest(&v1.ListGroupBuysRequest{
		PageSize:  2,
		PageToken: firstResp.Msg.NextPageToken,
	}))
	if err != nil {
		t.Fatalf("ListGroupBuys second page error: %v", err)
	}
	if len(secondResp.Msg.GroupBuys) != 1 {
		t.Fatalf("second page length = %d, want 1", len(secondResp.Msg.GroupBuys))
	}
	if secondResp.Msg.NextPageToken != "" {
		t.Fatalf("second page token = %q, want empty", secondResp.Msg.NextPageToken)
	}

	emptyResp, err := h.ListGroupBuys(context.Background(), connect.NewRequest(&v1.ListGroupBuysRequest{
		PageSize:  2,
		PageToken: "100",
	}))
	if err != nil {
		t.Fatalf("ListGroupBuys overflow page error: %v", err)
	}
	if len(emptyResp.Msg.GroupBuys) != 0 {
		t.Fatalf("overflow page length = %d, want 0", len(emptyResp.Msg.GroupBuys))
	}
	if emptyResp.Msg.NextPageToken != "" {
		t.Fatalf("overflow page token = %q, want empty", emptyResp.Msg.NextPageToken)
	}
}

func TestGroupBuyHandler_ListGroupBuys_InvalidPageToken(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	h := NewGroupBuyHandler(service.NewGroupBuyService(repo))

	_, err := h.ListGroupBuys(context.Background(), connect.NewRequest(&v1.ListGroupBuysRequest{
		PageSize:  10,
		PageToken: "bad-token",
	}))
	if err == nil {
		t.Fatal("expected invalid argument error, got nil")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("error code = %v, want %v", connect.CodeOf(err), connect.CodeInvalidArgument)
	}
}

func TestEventHandler_ListEvents_PaginationContract(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	now := time.Now()
	_ = repo.Create(context.Background(), &event.Event{
		ID:        "event-1",
		Title:     "Event-1",
		Status:    event.EventStatusActive,
		CreatedAt: now.Add(-1 * time.Hour),
	})
	_ = repo.Create(context.Background(), &event.Event{
		ID:        "event-2",
		Title:     "Event-2",
		Status:    event.EventStatusActive,
		CreatedAt: now.Add(-2 * time.Hour),
	})
	_ = repo.Create(context.Background(), &event.Event{
		ID:        "event-3",
		Title:     "Event-3",
		Status:    event.EventStatusActive,
		CreatedAt: now.Add(-3 * time.Hour),
	})

	firstResp, err := h.ListEvents(context.Background(), connect.NewRequest(&v1.ListEventsRequest{
		PageSize: 2,
	}))
	if err != nil {
		t.Fatalf("ListEvents first page error: %v", err)
	}
	if len(firstResp.Msg.Events) != 2 {
		t.Fatalf("first page length = %d, want 2", len(firstResp.Msg.Events))
	}
	if firstResp.Msg.NextPageToken != "2" {
		t.Fatalf("next page token = %q, want %q", firstResp.Msg.NextPageToken, "2")
	}

	secondResp, err := h.ListEvents(context.Background(), connect.NewRequest(&v1.ListEventsRequest{
		PageSize:  2,
		PageToken: firstResp.Msg.NextPageToken,
	}))
	if err != nil {
		t.Fatalf("ListEvents second page error: %v", err)
	}
	if len(secondResp.Msg.Events) != 1 {
		t.Fatalf("second page length = %d, want 1", len(secondResp.Msg.Events))
	}
	if secondResp.Msg.NextPageToken != "" {
		t.Fatalf("second page token = %q, want empty", secondResp.Msg.NextPageToken)
	}
}

func TestEventHandler_ListEvents_InvalidPageToken(t *testing.T) {
	repo := memory.NewEventRepository()
	h := NewEventHandler(service.NewEventService(repo))

	_, err := h.ListEvents(context.Background(), connect.NewRequest(&v1.ListEventsRequest{
		PageSize:  10,
		PageToken: "bad-token",
	}))
	if err == nil {
		t.Fatal("expected invalid argument error, got nil")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("error code = %v, want %v", connect.CodeOf(err), connect.CodeInvalidArgument)
	}
}
