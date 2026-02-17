package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateAndGetByID(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	u := &user.User{
		ID:        "user-1",
		Name:      "Alice",
		Email:     "alice@example.com",
		PhotoURL:  "https://example.com/photo.jpg",
		Role:      user.UserRoleUser,
		CreatedAt: time.Now().Truncate(time.Second),
		UpdatedAt: time.Now().Truncate(time.Second),
	}

	err := repo.Create(ctx, u)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "Alice", got.Name)
	assert.Equal(t, "alice@example.com", got.Email)
	assert.Equal(t, user.UserRoleUser, got.Role)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	u := &user.User{ID: "user-1", Name: "Bob", Email: "bob@example.com", Role: user.UserRoleCreator}
	require.NoError(t, repo.Create(ctx, u))

	got, err := repo.GetByEmail(ctx, "bob@example.com")
	require.NoError(t, err)
	assert.Equal(t, "user-1", got.ID)
	assert.Equal(t, user.UserRoleCreator, got.Role)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.GetByEmail(ctx, "nobody@example.com")
	assert.ErrorIs(t, err, user.ErrNotFound)
}

func TestUserRepository_Update(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	u := &user.User{ID: "user-1", Name: "Original", Email: "orig@example.com", Role: user.UserRoleUser}
	require.NoError(t, repo.Create(ctx, u))

	u.Name = "Updated"
	u.Role = user.UserRoleCreator
	require.NoError(t, repo.Update(ctx, u))

	got, err := repo.GetByID(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated", got.Name)
	assert.Equal(t, user.UserRoleCreator, got.Role)
}

func TestUserRepository_List(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		u := &user.User{
			ID:    fmt.Sprintf("user-%d", i),
			Name:  fmt.Sprintf("User %d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
		}
		require.NoError(t, repo.Create(ctx, u))
	}

	// Full list
	all, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, all, 5)

	// Pagination
	page, err := repo.List(ctx, 2, 0)
	require.NoError(t, err)
	assert.Len(t, page, 2)

	page2, err := repo.List(ctx, 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2, 2)
}
