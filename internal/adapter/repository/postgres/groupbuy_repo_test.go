package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

func TestGroupBuyRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	_, db := setupPostgresContainer(t)
	// Ensure groupbuy models are migrated
	err := db.AutoMigrate(
		&model.User{},
		&model.GroupBuy{},
		&model.Product{},
		&model.ProductSpec{},
		&model.Order{},
		&model.OrderItem{},
		&model.Category{},
	)
	require.NoError(t, err)

	userRepo := NewUserRepository(db)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	// Seed user data to satisfy foreign keys
	admin := &user.User{ID: "admin-1", Name: "Admin", Email: "admin1@test.com", Role: user.UserRoleSysAdmin}
	creator := &user.User{ID: "creator-1", Name: "Creator", Email: "creator1@test.com"}
	manager := &user.User{ID: "manager-1", Name: "Manager", Email: "manager1@test.com"}

	err = userRepo.Create(ctx, admin)
	require.NoError(t, err)
	err = userRepo.Create(ctx, creator)
	require.NoError(t, err)
	err = userRepo.Create(ctx, manager)
	require.NoError(t, err)

	t.Run("Create and Get GroupBuy", func(t *testing.T) {
		gb := &groupbuy.GroupBuy{
			ID:          "gb-1",
			Title:       "Test GB",
			Description: "Dest",
			CreatorID:   creator.ID,
			Managers:    []*user.User{manager}, // Use []*user.User instead of []string based on entity mapping/creation need initially
			Status:      groupbuy.GroupBuyStatusDraft,
			Products: []*groupbuy.Product{
				{
					ID:            "prod-1",
					Name:          "Test Product 1",
					PriceOriginal: 100, // Use PriceOriginal
					Specs: []*groupbuy.ProductSpec{
						{ID: "spec-1", Name: "Red"}, // No PriceAdjustment in domain model
					},
				},
			},
		}

		err := repo.Create(ctx, gb)
		assert.NoError(t, err)

		fetched, err := repo.GetByID(ctx, "gb-1")
		require.NoError(t, err)

		assert.Equal(t, "Test GB", fetched.Title)
		assert.Equal(t, creator.ID, fetched.CreatorID)
		assert.Len(t, fetched.Managers, 1)
		assert.Equal(t, manager.ID, fetched.Managers[0].ID)

		assert.Len(t, fetched.Products, 1)
		assert.Equal(t, "Test Product 1", fetched.Products[0].Name)
		assert.Len(t, fetched.Products[0].Specs, 1)
		assert.Equal(t, "Red", fetched.Products[0].Specs[0].Name)
	})

	t.Run("List GroupBuys - Public Flow", func(t *testing.T) {
		gbPublic := &groupbuy.GroupBuy{ID: "gb-2", Title: "Pub 1", CreatorID: creator.ID, Status: groupbuy.GroupBuyStatusActive}
		gbPublicEnded := &groupbuy.GroupBuy{ID: "gb-3", Title: "Pub 2", CreatorID: creator.ID, Status: groupbuy.GroupBuyStatusEnded}
		gbDraft := &groupbuy.GroupBuy{ID: "gb-4", Title: "Draft 1", CreatorID: creator.ID, Status: groupbuy.GroupBuyStatusDraft}

		repo.Create(ctx, gbPublic)
		repo.Create(ctx, gbPublicEnded)
		repo.Create(ctx, gbDraft)

		// Public anonymous user should see only StatusActive and StatusEnded
		publicList, err := repo.List(ctx, 10, 0, "", false, false)
		assert.NoError(t, err)
		assert.Len(t, publicList, 2)
		for _, g := range publicList {
			assert.Contains(t, []int{int(groupbuy.GroupBuyStatusActive), int(groupbuy.GroupBuyStatusEnded)}, int(g.Status))
		}

		// Sysadmin should see ALL
		adminList, err := repo.List(ctx, 10, 0, admin.ID, true, false)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(adminList), 4) // Including gb-1 created earlier

		// Creator should see their drafts and public groupbuys
		creatorList, err := repo.List(ctx, 10, 0, creator.ID, false, false)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(creatorList), 4) // All of them created by this Creator
	})

	t.Run("Update GroupBuy", func(t *testing.T) {
		gb, err := repo.GetByID(ctx, "gb-1")
		require.NoError(t, err)

		gb.Title = "Updated GB"
		gb.Description = "New Desc"
		// Change products
		gb.Products[0].Name = "Updated Product 1"

		err = repo.Update(ctx, gb)
		assert.NoError(t, err)

		fetched, err := repo.GetByID(ctx, "gb-1")
		assert.NoError(t, err)
		assert.Equal(t, "Updated GB", fetched.Title)
		assert.Equal(t, "Updated Product 1", fetched.Products[0].Name)
	})
}
