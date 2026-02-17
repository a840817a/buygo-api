package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/buygo/buygo-api/internal/domain/auth"
	"github.com/buygo/buygo-api/internal/domain/user"
)

type AuthService struct {
	userRepo     user.Repository
	authProvider auth.Provider
	tokenManager auth.TokenManager
}

func NewAuthService(userRepo user.Repository, authProvider auth.Provider, tokenManager auth.TokenManager) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		authProvider: authProvider,
		tokenManager: tokenManager,
	}
}

func (s *AuthService) LoginOrRegister(ctx context.Context, token string) (string, *user.User, error) {
	// 1. Verify Token
	info, err := s.authProvider.VerifyToken(ctx, token)
	if err != nil {
		return "", nil, fmt.Errorf("verify token: %w", err)
	}

	// 2. Check if user exists by ID (using Firebase UID as ID)
	userID := info.UID

	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			// 3. Register new user if not found
			now := time.Now()
			u = &user.User{
				ID:        userID,
				Name:      info.Name,
				Email:     info.Email,
				PhotoURL:  info.PhotoURL,
				Role:      user.UserRoleUser, // Default role
				CreatedAt: now,
				UpdatedAt: now,
			}

			// Dev/Mock Admin Check
			if token == "mock-token-admin" {
				u.Role = user.UserRoleSysAdmin
			}
			if err := s.userRepo.Create(ctx, u); err != nil {
				return "", nil, fmt.Errorf("create user: %w", err)
			}
		} else {
			return "", nil, fmt.Errorf("get user: %w", err)
		}
	} else {
		// User exists, optionally update
		// e.g. update photo if changed
	}

	// 4. Generate Session Token (JWT)
	accessToken, err := s.tokenManager.GenerateToken(u)
	if err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}

	return accessToken, u, nil
}

func (s *AuthService) GetMe(ctx context.Context, userID string) (*user.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) ListUsers(ctx context.Context, limit, offset int) ([]*user.User, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.userRepo.List(ctx, limit, offset)
}

func (s *AuthService) UpdateUserRole(ctx context.Context, userID string, role user.UserRole) (*user.User, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	u.Role = role
	u.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// ListAssignableManagers: Creator or Admin Only
func (s *AuthService) ListAssignableManagers(ctx context.Context, query string) ([]*user.User, error) {
	_, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	// Verify Permission
	if role != int(user.UserRoleCreator) && role != int(user.UserRoleSysAdmin) {
		return nil, ErrPermissionDenied
	}

	// Fetch users. Ideally Repo supports filtering by Role.
	// Since Repo interface List is generic, we might need to filter here or add Repo method.
	// For Memory Repo, filtering here is fine.
	// TODO: Add ListByRole to Repository interface for production optimization.
	users, err := s.userRepo.List(ctx, 1000, 0) // Fetch large batch
	if err != nil {
		return nil, err
	}

	var eligible []*user.User
	for _, u := range users {
		if u.Role == user.UserRoleCreator || u.Role == user.UserRoleSysAdmin {
			// Optional: Query Filter
			eligible = append(eligible, u)
		}
	}
	return eligible, nil
}
