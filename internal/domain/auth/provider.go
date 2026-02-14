package auth

import "context"

type TokenInfo struct {
	UID      string
	Email    string
	Name     string
	PhotoURL string
}

type Provider interface {
	VerifyToken(ctx context.Context, token string) (*TokenInfo, error)
}
