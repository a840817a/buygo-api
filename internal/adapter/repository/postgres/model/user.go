package model

import (
	"time"

	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

type User struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	Email     string `gorm:"uniqueIndex"`
	PhotoURL  string
	Role      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *User) ToDomain() *user.User {
	return &user.User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		PhotoURL:  u.PhotoURL,
		Role:      user.UserRole(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func FromDomainUser(u *user.User) *User {
	return &User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		PhotoURL:  u.PhotoURL,
		Role:      int(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
