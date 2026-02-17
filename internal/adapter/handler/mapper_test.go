package handler

import (
	"testing"
	"time"

	v1 "github.com/hatsubosi/buygo-api/api/v1"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"github.com/stretchr/testify/assert"
)

func TestToProtoUser_AllRoles(t *testing.T) {
	tests := []struct {
		name     string
		role     user.UserRole
		expected v1.UserRole
	}{
		{"User", user.UserRoleUser, v1.UserRole_USER_ROLE_USER},
		{"Creator", user.UserRoleCreator, v1.UserRole_USER_ROLE_CREATOR},
		{"SysAdmin", user.UserRoleSysAdmin, v1.UserRole_USER_ROLE_SYS_ADMIN},
		{"Unspecified", user.UserRoleUnspecified, v1.UserRole_USER_ROLE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &user.User{
				ID:       "uid-1",
				Name:     "Test",
				Email:    "t@example.com",
				PhotoURL: "http://photo.jpg",
				Role:     tt.role,
			}
			proto := toProtoUser(u)
			assert.Equal(t, "uid-1", proto.Id)
			assert.Equal(t, "Test", proto.Name)
			assert.Equal(t, "t@example.com", proto.Email)
			assert.Equal(t, "http://photo.jpg", proto.PhotoUrl)
			assert.Equal(t, tt.expected, proto.Role)
		})
	}
}

func TestToProtoUser_Nil(t *testing.T) {
	assert.Nil(t, toProtoUser(nil))
}

func TestToProtoGroupBuy(t *testing.T) {
	now := time.Now()
	deadline := now.Add(24 * time.Hour)

	gb := &groupbuy.GroupBuy{
		ID:             "gb-1",
		Title:          "My GroupBuy",
		Description:    "Desc",
		CoverImage:     "http://cover.jpg",
		Status:         groupbuy.GroupBuyStatusActive,
		CreatedAt:      now,
		Deadline:       &deadline,
		ExchangeRate:   0.25,
		Rounding:       &groupbuy.RoundingConfig{Method: 2, Digit: 1},
		SourceCurrency: "JPY",
		ShippingConfigs: []*groupbuy.ShippingConfig{
			{ID: "sc-1", Name: "Standard", Type: 1, Price: 100},
		},
	}

	proto := toProtoGroupBuy(gb)
	assert.Equal(t, "gb-1", proto.Id)
	assert.Equal(t, "My GroupBuy", proto.Title)
	assert.Equal(t, "Desc", proto.Description)
	assert.Equal(t, "http://cover.jpg", proto.CoverImageUrl)
	assert.Equal(t, v1.GroupBuyStatus(groupbuy.GroupBuyStatusActive), proto.Status)
	assert.Equal(t, 0.25, proto.ExchangeRate)
	assert.Equal(t, "JPY", proto.SourceCurrency)
	assert.NotNil(t, proto.Deadline)
	assert.NotNil(t, proto.RoundingConfig)
	assert.Equal(t, v1.RoundingMethod(2), proto.RoundingConfig.Method)
	assert.Equal(t, int32(1), proto.RoundingConfig.Digit)
	assert.Len(t, proto.ShippingConfigs, 1)
	assert.Equal(t, "sc-1", proto.ShippingConfigs[0].Id)
}

func TestToProtoGroupBuy_Nil(t *testing.T) {
	assert.Nil(t, toProtoGroupBuy(nil))
}

func TestToProtoGroupBuy_NoDeadline(t *testing.T) {
	gb := &groupbuy.GroupBuy{
		ID:       "gb-2",
		Rounding: &groupbuy.RoundingConfig{Method: 1, Digit: 0},
	}
	proto := toProtoGroupBuy(gb)
	assert.Nil(t, proto.Deadline)
}

func TestToProtoOrder(t *testing.T) {
	paidAt := time.Now()
	o := &groupbuy.Order{
		ID:               "ord-1",
		GroupBuyID:       "gb-1",
		UserID:           "user-1",
		TotalAmount:      5000,
		PaymentStatus:    groupbuy.PaymentStatusSubmitted,
		ContactInfo:      "John",
		ShippingAddress:  "123 Main St",
		ShippingMethodID: "sm-1",
		ShippingFee:      100,
		Note:             "Rush",
		Items: []*groupbuy.OrderItem{
			{
				ID:          "oi-1",
				ProductID:   "prod-1",
				SpecID:      "spec-1",
				Quantity:    3,
				Status:      groupbuy.OrderItemStatusUnordered,
				ProductName: "Widget",
				SpecName:    "Red",
				Price:       1500,
			},
		},
		PaymentInfo: &groupbuy.PaymentInfo{
			Method:       "Bank Transfer",
			AccountLast5: "12345",
			PaidAt:       &paidAt,
			Amount:       5000,
		},
	}

	proto := toProtoOrder(o)
	assert.Equal(t, "ord-1", proto.Id)
	assert.Equal(t, "gb-1", proto.GroupBuyId)
	assert.Equal(t, "user-1", proto.UserId)
	assert.Equal(t, int64(5000), proto.TotalAmount)
	assert.Equal(t, v1.PaymentStatus(2), proto.PaymentStatus)
	assert.Equal(t, "John", proto.ContactInfo)
	assert.Equal(t, "123 Main St", proto.ShippingAddress)
	assert.Equal(t, "sm-1", proto.ShippingMethodId)
	assert.Equal(t, int64(100), proto.ShippingFee)
	assert.Equal(t, "Rush", proto.Note)

	// Items
	assert.Len(t, proto.Items, 1)
	assert.Equal(t, "oi-1", proto.Items[0].Id)
	assert.Equal(t, "Widget", proto.Items[0].ProductName)
	assert.Equal(t, "Red", proto.Items[0].SpecName)
	assert.Equal(t, int32(3), proto.Items[0].Quantity)
	assert.Equal(t, int64(1500), proto.Items[0].Price)

	// Payment Info
	assert.NotNil(t, proto.PaymentInfo)
	assert.Equal(t, "Bank Transfer", proto.PaymentInfo.Method)
	assert.Equal(t, "12345", proto.PaymentInfo.AccountLast5)
	assert.NotNil(t, proto.PaymentInfo.PaidAt)
	assert.Equal(t, int64(5000), proto.PaymentInfo.Amount)
}

func TestToProtoOrder_Nil(t *testing.T) {
	assert.Nil(t, toProtoOrder(nil))
}

func TestToProtoOrder_NoPaymentInfo(t *testing.T) {
	o := &groupbuy.Order{
		ID:         "ord-2",
		GroupBuyID: "gb-1",
	}
	proto := toProtoOrder(o)
	assert.Nil(t, proto.PaymentInfo)
}

func TestToProtoProduct(t *testing.T) {
	p := &groupbuy.Product{
		ID:            "prod-1",
		GroupBuyID:    "gb-1",
		Name:          "Gadget",
		Description:   "Cool gadget",
		ImageURL:      "http://img.jpg",
		PriceOriginal: 1000,
		ExchangeRate:  0.25,
		Rounding:      &groupbuy.RoundingConfig{Method: 1, Digit: 0},
		PriceFinal:    250,
		MaxQuantity:   10,
		Specs: []*groupbuy.ProductSpec{
			{ID: "s1", Name: "Large"},
			{ID: "s2", Name: "Small"},
		},
	}

	proto := toProtoProduct(p)
	assert.Equal(t, "prod-1", proto.Id)
	assert.Equal(t, "Gadget", proto.Name)
	assert.Equal(t, int64(1000), proto.PriceOriginal)
	assert.Equal(t, int64(250), proto.PriceFinal)
	assert.Len(t, proto.Specs, 2)
	assert.Equal(t, "Large", proto.Specs[0].Name)
	assert.NotNil(t, proto.RoundingConfig)
}

func TestToProtoProduct_Nil(t *testing.T) {
	assert.Nil(t, toProtoProduct(nil))
}

func TestToProtoShippingConfig(t *testing.T) {
	c := &groupbuy.ShippingConfig{
		ID:    "sc-1",
		Name:  "Express",
		Type:  2,
		Price: 500,
	}
	proto := toProtoShippingConfig(c)
	assert.Equal(t, "sc-1", proto.Id)
	assert.Equal(t, "Express", proto.Name)
	assert.Equal(t, v1.ShippingType(2), proto.Type)
	assert.Equal(t, int64(500), proto.Price)
}

func TestToProtoShippingConfig_Nil(t *testing.T) {
	assert.Nil(t, toProtoShippingConfig(nil))
}

func TestFromProtoProduct(t *testing.T) {
	proto := &v1.Product{
		Id:            "prod-1",
		GroupBuyId:    "gb-1",
		Name:          "Widget",
		PriceOriginal: 500,
		ExchangeRate:  0.3,
		RoundingConfig: &v1.RoundingConfig{
			Method: v1.RoundingMethod(2),
			Digit:  1,
		},
		Specs: []*v1.ProductSpec{
			{Id: "s1", Name: "Red"},
		},
	}

	domain := fromProtoProduct(proto)
	assert.Equal(t, "prod-1", domain.ID)
	assert.Equal(t, "Widget", domain.Name)
	assert.Equal(t, int64(500), domain.PriceOriginal)
	assert.NotNil(t, domain.Rounding)
	assert.Equal(t, groupbuy.RoundingMethodCeil, domain.Rounding.Method)
	assert.Len(t, domain.Specs, 1)
}

func TestFromProtoProduct_Nil(t *testing.T) {
	assert.Nil(t, fromProtoProduct(nil))
}
