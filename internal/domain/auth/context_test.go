package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	role := 2 // user.UserRoleCreator

	t.Run("NewContext and FromContext", func(t *testing.T) {
		newCtx := NewContext(ctx, userID, role)

		gotID, gotRole, ok := FromContext(newCtx)
		assert.True(t, ok)
		assert.Equal(t, userID, gotID)
		assert.Equal(t, role, gotRole)
	})

	t.Run("FromContext with empty context", func(t *testing.T) {
		gotID, gotRole, ok := FromContext(ctx)
		assert.False(t, ok)
		assert.Empty(t, gotID)
		assert.Zero(t, gotRole)
	})

	t.Run("FromContext with partial context (only ID)", func(t *testing.T) {
		partialCtx := context.WithValue(ctx, userKey, userID)
		gotID, gotRole, ok := FromContext(partialCtx)
		assert.False(t, ok)
		assert.Equal(t, userID, gotID)
		assert.Zero(t, gotRole)
	})

	t.Run("FromContext with partial context (only Role)", func(t *testing.T) {
		partialCtx := context.WithValue(ctx, roleKey, role)
		gotID, gotRole, ok := FromContext(partialCtx)
		assert.False(t, ok)
		assert.Empty(t, gotID)
		assert.Equal(t, role, gotRole)
	})
}
