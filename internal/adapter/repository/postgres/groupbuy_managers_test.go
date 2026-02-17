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
