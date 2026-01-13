package service

import (
	"context"
	"fmt"

	"buygo/internal/adapter/auth"
	"buygo/internal/domain/user"
)

type AuthService struct {
	userRepo     user.Repository
	authProvider auth.Provider
}

func NewAuthService(userRepo user.Repository, authProvider auth.Provider) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		authProvider: authProvider,
	}
}

func (s *AuthService) LoginOrRegister(ctx context.Context, provider string, token string) (*user.User, error) {
	// 1. Verify Token
	info, err := s.authProvider.VerifyToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}

	// 2. Check if user exists by ProviderID (e.g. "firebase:uid")
	// For simplicity, using "firebase:" + info.UID
	providerID := "firebase:" + info.UID

	u, err := s.userRepo.GetByProviderID(ctx, providerID)
	if err == nil {
		// User exists, return logic
		return u, nil
	}

	// 3. Register new user if not found
	newUser := &user.User{
		ID: providerID, // Using providerID as ID for simplicity or generate UUID? Better generate UUID.
		// For now, let's use providerID as ID or we need a mapping.
		// Ideally: ID=UUID, ProviderID=...
		Name:       info.Name,
		Email:      info.Email,
		AvatarURL:  info.AvatarURL,
		ProviderID: providerID,
	}
	// TODO: Generate real UUID for ID
	newUser.ID = providerID // simplification for now

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}
