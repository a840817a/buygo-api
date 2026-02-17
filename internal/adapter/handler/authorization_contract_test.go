package handler

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"

	v1 "github.com/hatsubosi/buygo-api/api/v1"
	"github.com/hatsubosi/buygo-api/internal/adapter/repository/memory"
	domainAuth "github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"github.com/hatsubosi/buygo-api/internal/service"
)

func TestAuthHandler_ListUsers_AccessAndPaging(t *testing.T) {
	repo := memory.NewUserRepository()
	now := time.Now()
	_ = repo.Create(context.Background(), &user.User{
		ID:        "u-1",
		Name:      "User 1",
		Role:      user.UserRoleUser,
		CreatedAt: now,
		UpdatedAt: now,
	})
	_ = repo.Create(context.Background(), &user.User{
		ID:        "u-2",
		Name:      "User 2",
		Role:      user.UserRoleCreator,
		CreatedAt: now,
		UpdatedAt: now,
	})

	h := NewAuthHandler(service.NewAuthService(repo, nil, nil))

	_, err := h.ListUsers(context.Background(), connect.NewRequest(&v1.ListUsersRequest{}))
	if err == nil {
		t.Fatal("expected unauthenticated for anonymous request")
	}
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("anonymous code = %v, want %v", connect.CodeOf(err), connect.CodeUnauthenticated)
	}

	userCtx := domainAuth.NewContext(context.Background(), "u-1", int(user.UserRoleUser))
	_, err = h.ListUsers(userCtx, connect.NewRequest(&v1.ListUsersRequest{}))
	if err == nil {
		t.Fatal("expected permission denied for non-admin request")
	}
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("non-admin code = %v, want %v", connect.CodeOf(err), connect.CodePermissionDenied)
	}

	adminCtx := domainAuth.NewContext(context.Background(), "admin", int(user.UserRoleSysAdmin))
	_, err = h.ListUsers(adminCtx, connect.NewRequest(&v1.ListUsersRequest{
		PageSize:  10,
		PageToken: "bad",
	}))
	if err == nil {
		t.Fatal("expected invalid argument for bad page token")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("bad token code = %v, want %v", connect.CodeOf(err), connect.CodeInvalidArgument)
	}

	first, err := h.ListUsers(adminCtx, connect.NewRequest(&v1.ListUsersRequest{PageSize: 1}))
	if err != nil {
		t.Fatalf("first page error: %v", err)
	}
	if len(first.Msg.Users) != 1 {
		t.Fatalf("first page size = %d, want 1", len(first.Msg.Users))
	}
	if first.Msg.NextPageToken == "" {
		t.Fatal("expected next page token on first page")
	}

	second, err := h.ListUsers(adminCtx, connect.NewRequest(&v1.ListUsersRequest{
		PageSize:  1,
		PageToken: first.Msg.NextPageToken,
	}))
	if err != nil {
		t.Fatalf("second page error: %v", err)
	}
	if len(second.Msg.Users) != 1 {
		t.Fatalf("second page size = %d, want 1", len(second.Msg.Users))
	}
	if second.Msg.NextPageToken != "" {
		t.Fatalf("second page token = %q, want empty", second.Msg.NextPageToken)
	}
}

func TestAuthHandler_GetMe_Unauthenticated(t *testing.T) {
	h := NewAuthHandler(service.NewAuthService(memory.NewUserRepository(), nil, nil))

	_, err := h.GetMe(context.Background(), connect.NewRequest(&v1.GetMeRequest{}))
	if err == nil {
		t.Fatal("expected unauthenticated error")
	}
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("code = %v, want %v", connect.CodeOf(err), connect.CodeUnauthenticated)
	}
}

func TestAuthHandler_UpdateUserRole_InvalidRole(t *testing.T) {
	h := NewAuthHandler(service.NewAuthService(memory.NewUserRepository(), nil, nil))
	adminCtx := domainAuth.NewContext(context.Background(), "admin", int(user.UserRoleSysAdmin))

	_, err := h.UpdateUserRole(adminCtx, connect.NewRequest(&v1.UpdateUserRoleRequest{
		UserId: "u-1",
		Role:   v1.UserRole_USER_ROLE_UNSPECIFIED,
	}))
	if err == nil {
		t.Fatal("expected invalid role error")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want %v", connect.CodeOf(err), connect.CodeInvalidArgument)
	}
}

func TestGroupBuyHandler_UpdateGroupBuy_AccessControl(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := service.NewGroupBuyService(repo)
	h := NewGroupBuyHandler(svc)

	now := time.Now()
	_ = repo.Create(context.Background(), &groupbuy.GroupBuy{
		ID:         "gb-1",
		Title:      "Old Title",
		Status:     groupbuy.GroupBuyStatusDraft,
		CreatorID:  "creator-1",
		ManagerIDs: []string{"manager-1"},
		CreatedAt:  now,
	})

	_, err := h.UpdateGroupBuy(context.Background(), connect.NewRequest(&v1.UpdateGroupBuyRequest{
		GroupBuyId: "gb-1",
		Title:      "New Title",
	}))
	if err == nil {
		t.Fatal("expected unauthenticated for anonymous update")
	}
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("anonymous code = %v, want %v", connect.CodeOf(err), connect.CodeUnauthenticated)
	}

	otherCtx := domainAuth.NewContext(context.Background(), "other", int(user.UserRoleUser))
	_, err = h.UpdateGroupBuy(otherCtx, connect.NewRequest(&v1.UpdateGroupBuyRequest{
		GroupBuyId: "gb-1",
		Title:      "New Title",
	}))
	if err == nil {
		t.Fatal("expected permission denied for non-manager update")
	}
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("non-manager code = %v, want %v", connect.CodeOf(err), connect.CodePermissionDenied)
	}

	managerCtx := domainAuth.NewContext(context.Background(), "manager-1", int(user.UserRoleCreator))
	resp, err := h.UpdateGroupBuy(managerCtx, connect.NewRequest(&v1.UpdateGroupBuyRequest{
		GroupBuyId: "gb-1",
		Title:      "Manager Updated",
	}))
	if err != nil {
		t.Fatalf("manager update error: %v", err)
	}
	if resp.Msg.GroupBuy.GetTitle() != "Manager Updated" {
		t.Fatalf("title = %q, want %q", resp.Msg.GroupBuy.GetTitle(), "Manager Updated")
	}
}

func TestEventHandler_ListEventRegistrations_AccessControl(t *testing.T) {
	repo := memory.NewEventRepository()
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	now := time.Now()
	_ = repo.Create(context.Background(), &event.Event{
		ID:         "evt-1",
		Title:      "Event 1",
		Status:     event.EventStatusActive,
		CreatorID:  "creator-1",
		ManagerIDs: []string{"manager-1"},
		CreatedAt:  now,
		StartTime:  now.Add(24 * time.Hour),
		EndTime:    now.Add(48 * time.Hour),
	})

	_, err := h.ListEventRegistrations(context.Background(), connect.NewRequest(&v1.ListEventRegistrationsRequest{
		EventId: "evt-1",
	}))
	if err == nil {
		t.Fatal("expected unauthenticated for anonymous request")
	}
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("anonymous code = %v, want %v", connect.CodeOf(err), connect.CodeUnauthenticated)
	}

	otherCtx := domainAuth.NewContext(context.Background(), "other", int(user.UserRoleUser))
	_, err = h.ListEventRegistrations(otherCtx, connect.NewRequest(&v1.ListEventRegistrationsRequest{
		EventId: "evt-1",
	}))
	if err == nil {
		t.Fatal("expected permission denied for non-manager request")
	}
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("non-manager code = %v, want %v", connect.CodeOf(err), connect.CodePermissionDenied)
	}

	creatorCtx := domainAuth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	resp, err := h.ListEventRegistrations(creatorCtx, connect.NewRequest(&v1.ListEventRegistrationsRequest{
		EventId: "evt-1",
	}))
	if err != nil {
		t.Fatalf("creator request error: %v", err)
	}
	if len(resp.Msg.Registrations) != 0 {
		t.Fatalf("registrations len = %d, want 0", len(resp.Msg.Registrations))
	}
}

func TestEventHandler_ListManagerEvents_InvalidPageToken(t *testing.T) {
	h := NewEventHandler(service.NewEventService(memory.NewEventRepository()))
	ctx := domainAuth.NewContext(context.Background(), "manager-1", int(user.UserRoleCreator))

	_, err := h.ListManagerEvents(ctx, connect.NewRequest(&v1.ListManagerEventsRequest{
		PageSize:  20,
		PageToken: "-1",
	}))
	if err == nil {
		t.Fatal("expected invalid argument error for negative page token")
	}
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want %v", connect.CodeOf(err), connect.CodeInvalidArgument)
	}
}
