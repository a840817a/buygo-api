package auth

import (
	"context"
	"errors"
)

type TokenInfo struct {
	UID       string
	Email     string
	Name      string
	AvatarURL string
}

// Provider Port
type Provider interface {
	VerifyToken(ctx context.Context, token string) (*TokenInfo, error)
}

// Mock Provider
type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) VerifyToken(ctx context.Context, token string) (*TokenInfo, error) {
	if token == "invalid" {
		return nil, errors.New("invalid token")
	}
	// Mock success
	return &TokenInfo{
		UID:       "mock-user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "https://example.com/avatar.jpg",
	}, nil
}
