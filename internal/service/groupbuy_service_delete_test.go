package service

import (
	"context"
	"testing"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/memory"
	"github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupBuyService_DeleteProduct(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	creatorID := "creator-1"
	otherID := "other-user"
	gbID := "gb-1"
	prodID := "prod-1"

	// Prepare data
	ctx := context.Background()
	_ = repo.Create(ctx, &groupbuy.GroupBuy{
		ID:        gbID,
		CreatorID: creatorID,
		Status:    groupbuy.GroupBuyStatusDraft,
	})
	_ = repo.AddProduct(ctx, &groupbuy.Product{
		ID:         prodID,
		GroupBuyID: gbID,
		Name:       "Test Product",
	})

	t.Run("Unauthorized", func(t *testing.T) {
		err := svc.DeleteProduct(context.Background(), gbID, prodID)
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("PermissionDenied", func(t *testing.T) {
		otherCtx := auth.NewContext(context.Background(), otherID, int(user.UserRoleUser))
		err := svc.DeleteProduct(otherCtx, gbID, prodID)
		assert.ErrorIs(t, err, ErrPermissionDenied)
	})

	t.Run("Success", func(t *testing.T) {
		creatorCtx := auth.NewContext(context.Background(), creatorID, int(user.UserRoleCreator))
		err := svc.DeleteProduct(creatorCtx, gbID, prodID)
		require.NoError(t, err)

		// Verify deletion
		gb, _ := repo.GetByID(ctx, gbID)
		found := false
		for _, p := range gb.Products {
			if p.ID == prodID {
				found = true
				break
			}
		}
		assert.False(t, found)
	})

	t.Run("GroupBuyNotFound", func(t *testing.T) {
		creatorCtx := auth.NewContext(context.Background(), creatorID, int(user.UserRoleCreator))
		err := svc.DeleteProduct(creatorCtx, "nonexistent", prodID)
		assert.Error(t, err)
	})
}
