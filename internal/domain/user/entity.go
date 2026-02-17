package user

import (
	"context"
	"errors"
	"time"
)

// UserRole Enum
type UserRole int

const (
	UserRoleUnspecified UserRole = 0
	UserRoleUser        UserRole = 1
	UserRoleCreator     UserRole = 2
	UserRoleSysAdmin    UserRole = 3
)

// User Entity
type User struct {
	ID        string
	Name      string
	Email     string
	PhotoURL  string
	Role      UserRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *User) IsAdmin() bool {
	return u.Role == UserRoleSysAdmin
}

var (
	ErrNotFound = errors.New("user not found")
)

// Repository Port
type Repository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}
