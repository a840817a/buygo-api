package handler

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	v1 "github.com/hatsubosi/buygo-api/api/v1"
	"github.com/hatsubosi/buygo-api/internal/adapter/repository/memory"
	domainAuth "github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"github.com/hatsubosi/buygo-api/internal/service"
)

type mockProvider struct{}

func (m *mockProvider) VerifyToken(ctx context.Context, token string) (*domainAuth.TokenInfo, error) {
	return &domainAuth.TokenInfo{UID: token, Name: "Mock User", Email: "mock@example.com"}, nil
}

type mockTokenManager struct{}

func (m *mockTokenManager) GenerateToken(u *user.User) (string, error) {
	return "mock-access-token", nil
}
func (m *mockTokenManager) ParseToken(token string) (*domainAuth.Claims, error) {
	return nil, nil
}

func TestAuthHandler_Login(t *testing.T) {
	repo := memory.NewUserRepository()
	svc := service.NewAuthService(repo, &mockProvider{}, &mockTokenManager{})
	h := NewAuthHandler(svc)

	resp, err := h.Login(context.Background(), connect.NewRequest(&v1.LoginRequest{
		IdToken: "new-user-1",
	}))
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}

	if resp.Msg.User.Id != "new-user-1" {
		t.Errorf("got user id %q, want %q", resp.Msg.User.Id, "new-user-1")
	}
	if resp.Msg.AccessToken != "mock-access-token" {
		t.Errorf("got access token %q, want %q", resp.Msg.AccessToken, "mock-access-token")
	}
}

func TestAuthHandler_GetMe(t *testing.T) {
	repo := memory.NewUserRepository()
	_ = repo.Create(context.Background(), &user.User{
		ID:   "user-1",
		Name: "Test User",
		Role: user.UserRoleUser,
	})
	svc := service.NewAuthService(repo, nil, nil)
	h := NewAuthHandler(svc)

	ctx := domainAuth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	resp, err := h.GetMe(ctx, connect.NewRequest(&v1.GetMeRequest{}))
	if err != nil {
		t.Fatalf("GetMe error: %v", err)
	}

	if resp.Msg.User.Name != "Test User" {
		t.Errorf("got name %q, want %q", resp.Msg.User.Name, "Test User")
	}
}

func TestAuthHandler_ListAssignableManagers(t *testing.T) {
	repo := memory.NewUserRepository()
	_ = repo.Create(context.Background(), &user.User{ID: "admin", Role: user.UserRoleSysAdmin})
	_ = repo.Create(context.Background(), &user.User{ID: "creator", Role: user.UserRoleCreator})
	_ = repo.Create(context.Background(), &user.User{ID: "user", Role: user.UserRoleUser})

	svc := service.NewAuthService(repo, nil, nil)
	h := NewAuthHandler(svc)

	ctx := domainAuth.NewContext(context.Background(), "admin", int(user.UserRoleSysAdmin))
	resp, err := h.ListAssignableManagers(ctx, connect.NewRequest(&v1.ListAssignableManagersRequest{}))
	if err != nil {
		t.Fatalf("ListAssignableManagers error: %v", err)
	}

	// Should include admin and creator
	if len(resp.Msg.Managers) != 2 {
		t.Errorf("got %d users, want 2", len(resp.Msg.Managers))
	}
}
