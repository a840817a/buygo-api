package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

type JWTGenerator struct {
	secretKey []byte
	issuer    string
	expiry    time.Duration
}

func NewJWTGenerator(secretKey string, issuer string, expiry time.Duration) *JWTGenerator {
	return &JWTGenerator{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		expiry:    expiry,
	}
}

func (g *JWTGenerator) GenerateToken(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  u.ID,
		"iss":  g.issuer,
		"exp":  time.Now().Add(g.expiry).Unix(),
		"role": u.Role,
		"name": u.Name,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(g.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (g *JWTGenerator) ParseToken(tokenStr string) (*auth.Claims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sub, ok := claims["sub"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid subject in token")
		}

		var role user.UserRole
		if roleFloat, ok := claims["role"].(float64); ok {
			role = user.UserRole(roleFloat)
		} else {
			role = user.UserRoleUnspecified
		}

		return &auth.Claims{
			UserID: sub,
			Role:   role,
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}
