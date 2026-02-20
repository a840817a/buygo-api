package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

func setupPostgresContainer(t *testing.T) (*tcpostgres.PostgresContainer, *gorm.DB) {
	ctx := context.Background()

	dbName := "buygo_test"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		tcpostgres.WithDatabase(dbName),
		tcpostgres.WithUsername(dbUser),
		tcpostgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.User{})
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	return postgresContainer, db
}

func TestUserRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	_, db := setupPostgresContainer(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("Create and Get User", func(t *testing.T) {
		u := &user.User{
			ID:    "user-1",
			Name:  "Test User",
			Email: "test@example.com",
			Role:  user.UserRoleUser,
		}

		err := repo.Create(ctx, u)
		assert.NoError(t, err)

		fetched, err := repo.GetByID(ctx, "user-1")
		assert.NoError(t, err)
		assert.Equal(t, u.ID, fetched.ID)
		assert.Equal(t, u.Name, fetched.Name)
		assert.Equal(t, u.Email, fetched.Email)

		fetchedEmail, err := repo.GetByEmail(ctx, "test@example.com")
		assert.NoError(t, err)
		assert.Equal(t, u.ID, fetchedEmail.ID)
	})

	t.Run("Update User", func(t *testing.T) {
		u, err := repo.GetByID(ctx, "user-1")
		require.NoError(t, err)

		u.Name = "Updated Name"
		u.Role = user.UserRoleSysAdmin

		err = repo.Update(ctx, u)
		assert.NoError(t, err)

		fetched, err := repo.GetByID(ctx, "user-1")
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", fetched.Name)
		assert.True(t, fetched.IsAdmin())
	})

	t.Run("List Users", func(t *testing.T) {
		repo.Create(ctx, &user.User{ID: "user-2", Email: "test2@example.com"})
		repo.Create(ctx, &user.User{ID: "user-3", Email: "test3@example.com"})

		list, err := repo.List(ctx, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, list, 3)

		paged, err := repo.List(ctx, 2, 1)
		assert.NoError(t, err)
		assert.Len(t, paged, 2)
	})

	t.Run("Errors handling", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "non-existent")
		assert.ErrorIs(t, err, user.ErrNotFound)

		_, err = repo.GetByEmail(ctx, "nobody@example.com")
		assert.ErrorIs(t, err, user.ErrNotFound)
	})
}
