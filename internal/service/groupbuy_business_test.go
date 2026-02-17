package service

import (
	"context"
	"errors"
	"testing"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/memory"
	"github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ConfirmPayment ---

func TestConfirmPayment_ManagerOnly(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	anonCtx := context.Background()

	// Setup
	gb, _ := svc.CreateGroupBuy(creatorCtx, "GB", "Desc", nil, "", nil, nil, nil, 0, nil, "")
	svc.UpdateGroupBuy(creatorCtx, gb.ID, "", "", groupbuy.GroupBuyStatusActive, nil, "", nil, nil, nil, 0, nil, "")
	order, _ := svc.CreateOrder(userCtx, gb.ID, nil, "C", "A", "", "")

	// Anon → denied
	err := svc.ConfirmPayment(anonCtx, order.ID, 3)
	assert.True(t, errors.Is(err, ErrPermissionDenied), "Anon should not confirm payment")

	// User → denied
	err = svc.ConfirmPayment(userCtx, order.ID, 3)
	assert.True(t, errors.Is(err, ErrPermissionDenied), "User should not confirm payment")

	// Manager → success
	err = svc.ConfirmPayment(creatorCtx, order.ID, 3)
	assert.NoError(t, err)

	// SysAdmin → success
	adminCtx := auth.NewContext(context.Background(), "admin", int(user.UserRoleSysAdmin))
	err = svc.ConfirmPayment(adminCtx, order.ID, 2)
	assert.NoError(t, err)
}

// --- BatchUpdateStatus ---

func TestBatchUpdateStatus_AccessControl(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	anonCtx := context.Background()

	gb, _ := svc.CreateGroupBuy(creatorCtx, "GB", "Desc", nil, "", nil, nil, nil, 0, nil, "")

	// Anon → denied
	_, _, err := svc.BatchUpdateStatus(anonCtx, gb.ID, "", 2, 10)
	assert.True(t, errors.Is(err, ErrPermissionDenied))

	// User → denied
	_, _, err = svc.BatchUpdateStatus(userCtx, gb.ID, "", 2, 10)
	assert.True(t, errors.Is(err, ErrPermissionDenied))

	// Manager → success (memory repo stub returns 0,nil,nil)
	n, _, err := svc.BatchUpdateStatus(creatorCtx, gb.ID, "", 2, 10)
	assert.NoError(t, err)
	assert.Equal(t, int32(0), n) // Memory repo stub
}

func TestBatchUpdateStatus_InvalidTarget(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	gb, _ := svc.CreateGroupBuy(creatorCtx, "GB", "Desc", nil, "", nil, nil, nil, 0, nil, "")

	// Invalid target status (e.g. 99)
	_, _, err := svc.BatchUpdateStatus(creatorCtx, gb.ID, "", 99, 10)
	assert.Error(t, err, "Invalid target status should fail")

	// Target status 1 (Unordered) is not a valid batch target
	_, _, err = svc.BatchUpdateStatus(creatorCtx, gb.ID, "", 1, 10)
	assert.Error(t, err, "Target status 1 should fail")

	// Zero count → returns 0, nil
	n, ids, err := svc.BatchUpdateStatus(creatorCtx, gb.ID, "", 2, 0)
	assert.NoError(t, err)
	assert.Equal(t, int32(0), n)
	assert.Nil(t, ids)
}

func TestBatchUpdateStatus_ValidTransitions(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	gb, _ := svc.CreateGroupBuy(creatorCtx, "GB", "Desc", nil, "", nil, nil, nil, 0, nil, "")

	// All valid transitions should call repo without error
	validTargets := []int{2, 3, 4, 5, 6}
	for _, target := range validTargets {
		_, _, err := svc.BatchUpdateStatus(creatorCtx, gb.ID, "spec-1", target, 5)
		assert.NoError(t, err, "Target %d should be valid", target)
	}
}

// --- AddProduct ---

func TestAddProduct_AccessControl(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	anonCtx := context.Background()

	gb, _ := svc.CreateGroupBuy(creatorCtx, "GB", "Desc", nil, "", nil, nil, nil, 0, nil, "")

	// Anon → denied
	_, err := svc.AddProduct(anonCtx, gb.ID, "Prod", 100, 0, nil)
	assert.True(t, errors.Is(err, ErrPermissionDenied))

	// User → denied
	_, err = svc.AddProduct(userCtx, gb.ID, "Prod", 100, 0, nil)
	assert.True(t, errors.Is(err, ErrPermissionDenied))

	// Manager → success
	prod, err := svc.AddProduct(creatorCtx, gb.ID, "Widget", 1000, 0, []string{"Red", "Blue"})
	require.NoError(t, err)
	assert.Equal(t, "Widget", prod.Name)
	assert.NotEmpty(t, prod.ID)
}

func TestAddProduct_DefaultsFromGroupBuy(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	gb, _ := svc.CreateGroupBuy(creatorCtx, "GB", "Desc", nil, "", nil, nil, nil, 0, nil, "")

	// GroupBuy has default rate 0.23 and rounding Floor/Ones
	svc.UpdateGroupBuy(creatorCtx, gb.ID, "", "", 0, nil, "", nil, nil, nil, 0.23, &groupbuy.RoundingConfig{Method: 1, Digit: 0}, "")

	// AddProduct with rate=0 → should inherit from group buy
	prod, err := svc.AddProduct(creatorCtx, gb.ID, "Gadget", 100, 0, nil)
	require.NoError(t, err)
	assert.Equal(t, 0.23, prod.ExchangeRate)
	assert.Equal(t, int64(23), prod.PriceFinal) // 100 * 0.23 = 23 (Floor, Ones)
}

func TestAddProduct_SpecGeneration(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	gb, _ := svc.CreateGroupBuy(creatorCtx, "GB", "Desc", nil, "", nil, nil, nil, 0, nil, "")

	prod, err := svc.AddProduct(creatorCtx, gb.ID, "Item", 100, 1.0, []string{"S", "M", "", "L"})
	require.NoError(t, err)
	// Empty spec name "" should be skipped
	assert.Len(t, prod.Specs, 3)
	assert.Equal(t, "S", prod.Specs[0].Name)
	assert.Equal(t, "M", prod.Specs[1].Name)
	assert.Equal(t, "L", prod.Specs[2].Name)
	// Each spec should have product ID set
	for _, s := range prod.Specs {
		assert.Equal(t, prod.ID, s.ProductID)
		assert.NotEmpty(t, s.ID)
	}
}

// --- GetMyProjectOrder ---

func TestGetMyGroupBuyOrder(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userCtx := auth.NewContext(context.Background(), "user-1", int(user.UserRoleUser))
	anonCtx := context.Background()

	gb, _ := svc.CreateGroupBuy(creatorCtx, "GB", "Desc", nil, "", nil, nil, nil, 0, nil, "")
	svc.UpdateGroupBuy(creatorCtx, gb.ID, "", "", groupbuy.GroupBuyStatusActive, nil, "", nil, nil, nil, 0, nil, "")

	// Anon → denied
	_, err := svc.GetMyGroupBuyOrder(anonCtx, gb.ID)
	assert.True(t, errors.Is(err, ErrPermissionDenied))

	// No order → nil
	order, err := svc.GetMyGroupBuyOrder(userCtx, gb.ID)
	assert.NoError(t, err)
	assert.Nil(t, order)

	// Create order → returns it
	created, _ := svc.CreateOrder(userCtx, gb.ID, nil, "C", "A", "", "")
	order, err = svc.GetMyGroupBuyOrder(userCtx, gb.ID)
	assert.NoError(t, err)
	require.NotNil(t, order)
	assert.Equal(t, created.ID, order.ID)
}

// --- GetMyOrders ---

func TestGetMyOrders(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))
	userACtx := auth.NewContext(context.Background(), "user-a", int(user.UserRoleUser))
	userBCtx := auth.NewContext(context.Background(), "user-b", int(user.UserRoleUser))
	anonCtx := context.Background()

	gb, _ := svc.CreateGroupBuy(creatorCtx, "GB", "Desc", nil, "", nil, nil, nil, 0, nil, "")
	svc.UpdateGroupBuy(creatorCtx, gb.ID, "", "", groupbuy.GroupBuyStatusActive, nil, "", nil, nil, nil, 0, nil, "")

	// Anon → denied
	_, err := svc.GetMyOrders(anonCtx)
	assert.True(t, errors.Is(err, ErrPermissionDenied))

	// User A creates order
	svc.CreateOrder(userACtx, gb.ID, nil, "C", "A", "", "")
	// User B creates order
	svc.CreateOrder(userBCtx, gb.ID, nil, "C", "A", "", "")

	// Each user sees only own orders
	ordersA, err := svc.GetMyOrders(userACtx)
	assert.NoError(t, err)
	assert.Len(t, ordersA, 1)

	ordersB, err := svc.GetMyOrders(userBCtx)
	assert.NoError(t, err)
	assert.Len(t, ordersB, 1)

	assert.NotEqual(t, ordersA[0].ID, ordersB[0].ID)
}
