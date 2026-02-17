package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupBuyRepository_CreateWithManagers(t *testing.T) {
	db := newTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	// Create users
	creatorID := "user-creator"
	managerID := "user-manager"
	err := db.Create(&model.User{ID: creatorID, Name: "Creator", Email: "creator@test.com"}).Error
	require.NoError(t, err)
	err = db.Create(&model.User{ID: managerID, Name: "Manager", Email: "manager@test.com"}).Error
	require.NoError(t, err)

	gb := &groupbuy.GroupBuy{
		ID:         "gb-managers",
		Title:      "Group Buy With Managers",
		Status:     groupbuy.GroupBuyStatusActive,
		CreatorID:  creatorID,
		CreatedAt:  time.Now(),
		ManagerIDs: []string{managerID}, // Set Manager ID
	}

	// Test Create
	err = repo.Create(ctx, gb)
	require.NoError(t, err)

	// Fetch and Verify
	saved, err := repo.GetByID(ctx, "gb-managers")
	require.NoError(t, err)

	require.Len(t, saved.ManagerIDs, 1, "ManagerIDs should have 1 element")
	assert.Equal(t, managerID, saved.ManagerIDs[0])
	require.Len(t, saved.Managers, 1, "Managers should have 1 element")
	assert.Equal(t, "Manager", saved.Managers[0].Name)
}

func TestGroupBuyRepository_UpdateManagers(t *testing.T) {
	db := newTestDB(t)
	repo := NewGroupBuyRepository(db)
	ctx := context.Background()

	creatorID := "user-creator"
	managerA := "user-manager-a"
	managerB := "user-manager-b"

	require.NoError(t, db.Create(&model.User{ID: creatorID, Name: "Creator", Email: "creator@test.com"}).Error)
	require.NoError(t, db.Create(&model.User{ID: managerA, Name: "Manager A", Email: "manager-a@test.com"}).Error)
	require.NoError(t, db.Create(&model.User{ID: managerB, Name: "Manager B", Email: "manager-b@test.com"}).Error)

	gb := &groupbuy.GroupBuy{
		ID:         "gb-update-managers",
		Title:      "Group Buy Update Managers",
		Status:     groupbuy.GroupBuyStatusActive,
		CreatorID:  creatorID,
		CreatedAt:  time.Now(),
		ManagerIDs: []string{managerA},
	}
	require.NoError(t, repo.Create(ctx, gb))

	gb.ManagerIDs = []string{managerB}
	require.NoError(t, repo.Update(ctx, gb))

	saved, err := repo.GetByID(ctx, gb.ID)
	require.NoError(t, err)
	require.Len(t, saved.ManagerIDs, 1)
	assert.Equal(t, managerB, saved.ManagerIDs[0])
	require.Len(t, saved.Managers, 1)
	assert.Equal(t, managerB, saved.Managers[0].ID)
}
