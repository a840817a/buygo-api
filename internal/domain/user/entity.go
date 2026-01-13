package user

import (
	"context"
	"errors"
)

// User Entity
type User struct {
	ID            string
	Name          string
	Email         string
	AvatarURL     string
	ProviderID    string // e.g., "google:12345"
	IsSystemAdmin bool
}

var (
	ErrNotFound = errors.New("user not found")
)

// Repository Port
type Repository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByProviderID(ctx context.Context, providerID string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
}

// Service Port (UseCase)
type Service interface {
	LoginOrRegister(ctx context.Context, provider string, token string) (*User, error)
	GetProfile(ctx context.Context, id string) (*User, error)
}
