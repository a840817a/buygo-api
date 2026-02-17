package auth

import (
	"testing"
	"time"

	domainAuth "github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT_RoundTrip(t *testing.T) {
	gen := NewJWTGenerator("test-secret", "buygo-test", 1*time.Hour)

	u := &user.User{
		ID:   "user-123",
		Name: "Test User",
		Role: user.UserRoleCreator,
	}

	token, err := gen.GenerateToken(u)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := gen.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, user.UserRoleCreator, claims.Role)
}

func TestJWT_AllRoles(t *testing.T) {
	gen := NewJWTGenerator("test-secret", "buygo-test", 1*time.Hour)

	tests := []struct {
		name string
		role user.UserRole
	}{
		{"User", user.UserRoleUser},
		{"Creator", user.UserRoleCreator},
		{"SysAdmin", user.UserRoleSysAdmin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &user.User{ID: "uid", Name: "N", Role: tt.role}
			token, err := gen.GenerateToken(u)
			require.NoError(t, err)

			claims, err := gen.ParseToken(token)
			require.NoError(t, err)
			assert.Equal(t, tt.role, claims.Role)
		})
	}
}

func TestJWT_WrongKey(t *testing.T) {
	genA := NewJWTGenerator("key-A", "buygo", 1*time.Hour)
	genB := NewJWTGenerator("key-B", "buygo", 1*time.Hour)

	u := &user.User{ID: "uid", Name: "N", Role: user.UserRoleUser}
	token, err := genA.GenerateToken(u)
	require.NoError(t, err)

	_, err = genB.ParseToken(token)
	assert.Error(t, err)
}

func TestJWT_InvalidTokenString(t *testing.T) {
	gen := NewJWTGenerator("secret", "buygo", 1*time.Hour)

	_, err := gen.ParseToken("not-a-jwt")
	assert.Error(t, err)

	_, err = gen.ParseToken("")
	assert.Error(t, err)
}

func TestJWT_ExpiredToken(t *testing.T) {
	// Create a generator with negative expiry → token is already expired
	gen := NewJWTGenerator("secret", "buygo", -1*time.Hour)

	u := &user.User{ID: "uid", Name: "N", Role: user.UserRoleUser}
	token, err := gen.GenerateToken(u)
	require.NoError(t, err)

	_, err = gen.ParseToken(token)
	assert.Error(t, err, "expired token should fail parsing")
}

// Verify the Claims struct matches the domain interface
func TestJWT_ImplementsTokenManager(t *testing.T) {
	var _ domainAuth.TokenManager = (*JWTGenerator)(nil)
}
