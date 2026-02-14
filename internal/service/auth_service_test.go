package service

import (
	"context"
	"testing"

	"github.com/buygo/buygo-api/internal/adapter/repository/memory"
	"github.com/buygo/buygo-api/internal/domain/auth"
	"github.com/buygo/buygo-api/internal/domain/user"
)

// Mock Provider
type MockAuthProvider struct{}

func (p *MockAuthProvider) VerifyToken(ctx context.Context, token string) (*auth.TokenInfo, error) {
	if token == "valid-token" {
		return &auth.TokenInfo{
			UID:      "test-uid",
			Email:    "test@example.com",
			Name:     "Test User",
			PhotoURL: "http://example.com/photo.jpg",
		}, nil
	}
	return nil, context.DeadlineExceeded // Just an error
}

// Mock Token Generator
type MockTokenGen struct{}

func (g *MockTokenGen) GenerateToken(u *user.User) (string, error) {
	return "mock-jwt-token", nil
}
func (g *MockTokenGen) ParseToken(token string) (*auth.Claims, error) {
	return &auth.Claims{UserID: "test-uid", Role: user.UserRoleUser}, nil
}

func TestLoginOrRegister(t *testing.T) {
	repo := memory.NewUserRepository()
	provider := &MockAuthProvider{}
	tokenGen := &MockTokenGen{}

	svc := NewAuthService(repo, provider, tokenGen)

	// Case 1: Register New User
	token, u, err := svc.LoginOrRegister(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token != "mock-jwt-token" {
		t.Errorf("expected token mock-jwt-token, got %s", token)
	}
	if u.ID != "test-uid" {
		t.Errorf("expected user id test-uid, got %s", u.ID)
	}

	// Verify user is in repo
	savedUser, err := repo.GetByID(context.Background(), "test-uid")
	if err != nil {
		t.Fatalf("expected user to be saved, got error %v", err)
	}
	if savedUser.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", savedUser.Email)
	}

	// Case 2: Login Existing User
	token2, u2, err := svc.LoginOrRegister(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("expected no error on login, got %v", err)
	}
	if u2.ID != u.ID {
		t.Errorf("expected same user ID, got %s and %s", u2.ID, u.ID)
	}
	if token2 != "mock-jwt-token" {
		t.Errorf("expected token mock-jwt-token, got %s", token2)
	}
}

func TestLoginOrRegister_InvalidToken(t *testing.T) {
	repo := memory.NewUserRepository()
	provider := &MockAuthProvider{}
	tokenGen := &MockTokenGen{}

	svc := NewAuthService(repo, provider, tokenGen)

	_, _, err := svc.LoginOrRegister(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}

func TestUpdateUserRole(t *testing.T) {
	repo := memory.NewUserRepository()
	provider := &MockAuthProvider{}
	tokenGen := &MockTokenGen{}
	svc := NewAuthService(repo, provider, tokenGen)

	// Register a user first
	_, u, err := svc.LoginOrRegister(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if u.Role != user.UserRoleUser {
		t.Fatalf("expected default role User, got %d", u.Role)
	}

	// Update to Creator
	updated, err := svc.UpdateUserRole(context.Background(), u.ID, user.UserRoleCreator)
	if err != nil {
		t.Fatalf("UpdateUserRole failed: %v", err)
	}
	if updated.Role != user.UserRoleCreator {
		t.Errorf("expected role Creator, got %d", updated.Role)
	}

	// Verify persistence
	fetched, _ := svc.GetMe(context.Background(), u.ID)
	if fetched.Role != user.UserRoleCreator {
		t.Errorf("persisted role should be Creator, got %d", fetched.Role)
	}

	// Update non-existent user
	_, err = svc.UpdateUserRole(context.Background(), "non-existent", user.UserRoleSysAdmin)
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestListUsers(t *testing.T) {
	repo := memory.NewUserRepository()
	provider := &MockAuthProvider{}
	tokenGen := &MockTokenGen{}
	svc := NewAuthService(repo, provider, tokenGen)

	// Register a user
	svc.LoginOrRegister(context.Background(), "valid-token")

	// Default limit (0 → 10)
	users, err := svc.ListUsers(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}

	// Explicit limit
	users, err = svc.ListUsers(context.Background(), 5, 0)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
}

func TestListAssignableManagers(t *testing.T) {
	repo := memory.NewUserRepository()
	provider := &MockAuthProvider{}
	tokenGen := &MockTokenGen{}
	svc := NewAuthService(repo, provider, tokenGen)

	// Seed users
	repo.Create(context.Background(), &user.User{ID: "admin-1", Role: user.UserRoleSysAdmin})
	repo.Create(context.Background(), &user.User{ID: "creator-1", Role: user.UserRoleCreator})
	repo.Create(context.Background(), &user.User{ID: "user-1", Role: user.UserRoleUser})

	// User → permission denied
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	_, err := svc.ListAssignableManagers(userCtx, "")
	if err == nil {
		t.Error("Regular user should not list managers")
	}

	// Creator → success, returns Creator + Admin only
	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	managers, err := svc.ListAssignableManagers(creatorCtx, "")
	if err != nil {
		t.Fatalf("Creator ListAssignableManagers failed: %v", err)
	}
	if len(managers) != 2 {
		t.Errorf("expected 2 managers (admin + creator), got %d", len(managers))
	}

	// Admin → success
	adminCtx := auth.NewContext(context.Background(), "admin-1", int(user.UserRoleSysAdmin))
	managers, err = svc.ListAssignableManagers(adminCtx, "")
	if err != nil {
		t.Fatalf("Admin ListAssignableManagers failed: %v", err)
	}
	if len(managers) != 2 {
		t.Errorf("expected 2 managers, got %d", len(managers))
	}

	// Anon context → permission denied
	_, err = svc.ListAssignableManagers(context.Background(), "")
	if err == nil {
		t.Error("Anon should not list managers")
	}
}
