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

func TestGroupBuyService_PriceTemplate_AccessControl(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	adminCtx := auth.NewContext(context.Background(), "admin", int(user.UserRoleSysAdmin))
	creatorCtx := auth.NewContext(context.Background(), "creator", int(user.UserRoleCreator))
	anonCtx := context.Background()

	// Create: Admin Only
	t.Run("Create Access", func(t *testing.T) {
		_, err := svc.CreatePriceTemplate(anonCtx, "T1", "USD", 1.0, nil)
		assert.ErrorIs(t, err, ErrPermissionDenied)

		_, err = svc.CreatePriceTemplate(creatorCtx, "T1", "USD", 1.0, nil)
		assert.ErrorIs(t, err, ErrPermissionDenied)

		pt, err := svc.CreatePriceTemplate(adminCtx, "T1", "USD", 1.0, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, pt.ID)
	})

	// List: Authenticated
	t.Run("List Access", func(t *testing.T) {
		_, err := svc.ListPriceTemplates(anonCtx)
		assert.ErrorIs(t, err, ErrPermissionDenied)

		list, err := svc.ListPriceTemplates(creatorCtx)
		require.NoError(t, err)
		assert.Len(t, list, 1) // One created via admin above
	})

	// Update: Admin Only
	t.Run("Update Access", func(t *testing.T) {
		pt, _ := svc.CreatePriceTemplate(adminCtx, "ToUpdate", "USD", 1.0, nil)

		_, err := svc.UpdatePriceTemplate(anonCtx, pt.ID, "NewName", "", 0, nil)
		assert.ErrorIs(t, err, ErrPermissionDenied)

		_, err = svc.UpdatePriceTemplate(creatorCtx, pt.ID, "NewName", "", 0, nil)
		assert.ErrorIs(t, err, ErrPermissionDenied)

		updated, err := svc.UpdatePriceTemplate(adminCtx, pt.ID, "NewName", "", 2.0, nil)
		require.NoError(t, err)
		assert.Equal(t, "NewName", updated.Name)
		assert.Equal(t, 2.0, updated.ExchangeRate)
	})

	// Delete: Admin Only
	t.Run("Delete Access", func(t *testing.T) {
		pt, _ := svc.CreatePriceTemplate(adminCtx, "ToDelete", "USD", 1.0, nil)

		err := svc.DeletePriceTemplate(anonCtx, pt.ID)
		assert.ErrorIs(t, err, ErrPermissionDenied)

		err = svc.DeletePriceTemplate(creatorCtx, pt.ID)
		assert.ErrorIs(t, err, ErrPermissionDenied)

		err = svc.DeletePriceTemplate(adminCtx, pt.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = svc.GetPriceTemplate(adminCtx, pt.ID)
		assert.Error(t, err) // Expect not found
	})
}

func TestGroupBuyService_PriceTemplate_CRUD(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)
	ctx := auth.NewContext(context.Background(), "admin", int(user.UserRoleSysAdmin))

	// 1. Create
	rounding := &groupbuy.RoundingConfig{Method: groupbuy.RoundingMethodCeil, Digit: 1}
	pt, err := svc.CreatePriceTemplate(ctx, "Standard JPY", "JPY", 0.25, rounding)
	require.NoError(t, err)
	assert.Equal(t, "Standard JPY", pt.Name)
	assert.Equal(t, "JPY", pt.SourceCurrency)
	assert.Equal(t, 0.25, pt.ExchangeRate)
	assert.Equal(t, groupbuy.RoundingMethodCeil, pt.Rounding.Method)

	// 2. Get
	fetched, err := svc.GetPriceTemplate(ctx, pt.ID)
	require.NoError(t, err)
	assert.Equal(t, pt.ID, fetched.ID)
	assert.Equal(t, pt.Name, fetched.Name)

	// 3. List
	pt2, _ := svc.CreatePriceTemplate(ctx, "Another", "USD", 1.0, nil)
	list, err := svc.ListPriceTemplates(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Contains(t, []string{list[0].ID, list[1].ID}, pt.ID)
	assert.Contains(t, []string{list[0].ID, list[1].ID}, pt2.ID)

	// 4. Update
	updated, err := svc.UpdatePriceTemplate(ctx, pt.ID, "Updated JPY", "", 0.26, nil)
	require.NoError(t, err)
	assert.Equal(t, "Updated JPY", updated.Name)
	assert.Equal(t, 0.26, updated.ExchangeRate)
	assert.Equal(t, "JPY", updated.SourceCurrency) // Unchanged
	// Rounding should remain if nil passed?
	// The implementation checks `if rounding != nil { pt.Rounding = rounding }`
	// So passing nil should keep old rounding.
	assert.Equal(t, groupbuy.RoundingMethodCeil, updated.Rounding.Method)

	// 5. Delete
	err = svc.DeletePriceTemplate(ctx, pt.ID)
	require.NoError(t, err)

	listAfter, err := svc.ListPriceTemplates(ctx)
	require.NoError(t, err)
	assert.Len(t, listAfter, 1)
	assert.Equal(t, pt2.ID, listAfter[0].ID)
}
