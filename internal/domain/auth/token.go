package auth

import "github.com/hatsubosi/buygo-api/internal/domain/user"

type TokenManager interface {
	GenerateToken(user *user.User) (string, error)
	ParseToken(token string) (*Claims, error)
}

type Claims struct {
	UserID string
	Role   user.UserRole
}
