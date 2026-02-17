package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserRole(t *testing.T) {
	tests := []struct {
		role UserRole
		val  int
	}{
		{UserRoleUnspecified, 0},
		{UserRoleUser, 1},
		{UserRoleCreator, 2},
		{UserRoleSysAdmin, 3},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.val, int(tt.role))
	}
}

func TestUser(t *testing.T) {
	u := &User{
		ID:    "user-1",
		Name:  "Test User",
		Email: "test@example.com",
		Role:  UserRoleUser,
	}

	assert.Equal(t, "user-1", u.ID)
	assert.Equal(t, "Test User", u.Name)
	assert.Equal(t, "test@example.com", u.Email)
	assert.Equal(t, UserRoleUser, u.Role)
	assert.False(t, u.IsAdmin())

	u.Role = UserRoleSysAdmin
	assert.True(t, u.IsAdmin())
}
