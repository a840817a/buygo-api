package service

import (
	"context"
	"testing"
	"time"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/memory"
	"github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"github.com/stretchr/testify/assert"
)

func TestGroupBuyService_CreateGroupBuy_FullPayload(t *testing.T) {
	repo := memory.NewGroupBuyRepository()
	svc := NewGroupBuyService(repo)

	creatorCtx := auth.NewContext(context.Background(), "creator-1", int(user.UserRoleCreator))

	// Define full payload
	title := "Full Group Buy"
	desc := "With products and settings"
	products := []*groupbuy.Product{
		{
			Name:          "Product A",
			PriceOriginal: 1000,
			ExchangeRate:  0.25, // Explicit rate
		},
		{
			Name:          "Product B",
			PriceOriginal: 2000,
			// No rate, should use GB default
		},
	}
	rounding := &groupbuy.RoundingConfig{Method: groupbuy.RoundingMethodCeil, Digit: 1}
	deadline := time.Now().Add(24 * time.Hour)
	shippingConfigs := []*groupbuy.ShippingConfig{
		{Name: "Delivery", Price: 500},
	}
	managerIDs := []string{"manager-2"}

	// CALL SERVICE - This signature will change
	// Intentionally defining variables to pass to the new signature
	exchangeRate := 0.23
	coverImage := "http://cover.jpg"
	sourceCurrency := "USD"

	gb, err := svc.CreateGroupBuy(creatorCtx, title, desc, products, coverImage, &deadline, shippingConfigs, managerIDs, exchangeRate, rounding, sourceCurrency)

	assert.NoError(t, err)
	assert.NotNil(t, gb)
	assert.Equal(t, title, gb.Title)
	assert.Equal(t, desc, gb.Description)
	assert.Equal(t, groupbuy.GroupBuyStatusDraft, gb.Status)
	assert.Equal(t, exchangeRate, gb.ExchangeRate)
	assert.Equal(t, sourceCurrency, gb.SourceCurrency)
	assert.Equal(t, rounding, gb.Rounding)
	assert.WithinDuration(t, deadline, *gb.Deadline, time.Second)
	assert.Equal(t, "http://cover.jpg", gb.CoverImage)

	// Check Managers
	assert.Contains(t, gb.ManagerIDs, "creator-1") // Creator always added
	assert.Contains(t, gb.ManagerIDs, "manager-2") // Additional manager

	// Check Shipping
	assert.Len(t, gb.ShippingConfigs, 1)
	assert.Equal(t, "Delivery", gb.ShippingConfigs[0].Name)

	// Check Products
	assert.Len(t, gb.Products, 2)

	pA := gb.Products[0]
	assert.Equal(t, "Product A", pA.Name)
	assert.Equal(t, 0.25, pA.ExchangeRate)
	// Price calculation check: 1000 * 0.25 = 250. Rounding Ceil, Digit 1 (Tens). 250 -> 250.

	pB := gb.Products[1]
	assert.Equal(t, "Product B", pB.Name)
	assert.Equal(t, 0.23, pB.ExchangeRate) // Inherited
}
