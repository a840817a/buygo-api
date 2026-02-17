package model

import (
	"testing"
	"time"

	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"github.com/stretchr/testify/assert"
)

func TestUserModel_Conversion(t *testing.T) {
	now := time.Now()
	domainUser := &user.User{
		ID:        "u1",
		Name:      "Alice",
		Email:     "alice@example.com",
		PhotoURL:  "http://pic",
		Role:      user.UserRoleCreator,
		CreatedAt: now,
		UpdatedAt: now,
	}

	modelUser := FromDomainUser(domainUser)
	assert.Equal(t, domainUser.ID, modelUser.ID)
	assert.Equal(t, int(domainUser.Role), modelUser.Role)

	convertedBack := modelUser.ToDomain()
	assert.Equal(t, domainUser, convertedBack)
}

func TestGroupBuyModel_Conversion(t *testing.T) {
	now := time.Now()
	deadline := now.Add(24 * time.Hour)
	gb := &groupbuy.GroupBuy{
		ID:             "gb1",
		Title:          "Title",
		Description:    "Desc",
		CoverImage:     "http://cover",
		Status:         groupbuy.GroupBuyStatusActive,
		ExchangeRate:   31.5,
		SourceCurrency: "USD",
		CreatedAt:      now,
		Deadline:       &deadline,
		CreatorID:      "c1",
		ManagerIDs:     []string{"m1", "m2"},
		ShippingConfigs: []*groupbuy.ShippingConfig{
			{ID: "s1", Name: "Ship 1", Type: groupbuy.ShippingTypeDelivery, Price: 100},
		},
	}

	m := FromDomainGroupBuy(gb)
	assert.Equal(t, gb.ID, m.ID)
	assert.Equal(t, int(gb.Status), m.Status)

	converted := m.ToDomain()
	assert.Equal(t, gb.ID, converted.ID)
	assert.Equal(t, gb.Status, converted.Status)
	assert.Equal(t, gb.ExchangeRate, converted.ExchangeRate)
	assert.Equal(t, gb.SourceCurrency, converted.SourceCurrency)
	assert.Len(t, converted.ManagerIDs, 2)
	assert.Len(t, converted.ShippingConfigs, 1)
}

func TestProductModel_Conversion(t *testing.T) {
	p := &groupbuy.Product{
		ID:            "p1",
		GroupBuyID:    "gb1",
		Name:          "Product 1",
		PriceOriginal: 1000,
		PriceFinal:    31500,
		Specs: []*groupbuy.ProductSpec{
			{ID: "spec1", Name: "Spec 1"},
		},
	}

	m := FromDomainProduct(p)
	assert.Equal(t, p.ID, m.ID)

	converted := m.ToDomain()
	assert.Equal(t, p.ID, converted.ID)
	assert.Len(t, converted.Specs, 1)
	assert.Equal(t, "spec1", converted.Specs[0].ID)
}
