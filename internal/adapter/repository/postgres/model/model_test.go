package model

import (
	"testing"
	"time"

	"github.com/hatsubosi/buygo-api/internal/domain/event"
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

func TestCategoryModel_Conversion(t *testing.T) {
	cat := &groupbuy.Category{
		ID:        "cat1",
		Name:      "Electronics",
		SpecNames: []string{"Color", "Size"},
	}

	m := FromDomainCategory(cat)
	assert.Equal(t, cat.ID, m.ID)
	assert.Equal(t, cat.Name, m.Name)
	assert.Equal(t, cat.SpecNames, []string(m.SpecNames))

	converted := m.ToDomain()
	assert.Equal(t, cat.ID, converted.ID)
	assert.Equal(t, cat.Name, converted.Name)
	assert.Equal(t, cat.SpecNames, converted.SpecNames)
}

func TestPriceTemplateModel_Conversion(t *testing.T) {
	pt := &groupbuy.PriceTemplate{
		ID:             "pt1",
		Name:           "JPY Template",
		SourceCurrency: "JPY",
		ExchangeRate:   0.22,
		Rounding:       &groupbuy.RoundingConfig{Method: groupbuy.RoundingMethodRound, Digit: 0},
	}

	m := FromDomainPriceTemplate(pt)
	assert.Equal(t, pt.ID, m.ID)
	assert.Equal(t, pt.Name, m.Name)
	assert.Equal(t, pt.SourceCurrency, m.SourceCurrency)
	assert.Equal(t, pt.ExchangeRate, m.ExchangeRate)
	assert.Equal(t, int(groupbuy.RoundingMethodRound), m.RoundingMethod)
	assert.Equal(t, 0, m.RoundingDigit)

	converted := m.ToDomain()
	assert.Equal(t, pt.ID, converted.ID)
	assert.Equal(t, pt.Name, converted.Name)
	assert.Equal(t, pt.ExchangeRate, converted.ExchangeRate)
	assert.Equal(t, pt.Rounding.Method, converted.Rounding.Method)
	assert.Equal(t, pt.Rounding.Digit, converted.Rounding.Digit)
}

func TestPriceTemplateModel_NilHandling(t *testing.T) {
	assert.Nil(t, FromDomainPriceTemplate(nil))

	var m *PriceTemplate
	assert.Nil(t, m.ToDomain())
}

func TestPriceTemplateModel_NilRounding(t *testing.T) {
	pt := &groupbuy.PriceTemplate{
		ID:             "pt2",
		Name:           "No Rounding",
		SourceCurrency: "USD",
		ExchangeRate:   31.5,
	}
	m := FromDomainPriceTemplate(pt)
	assert.Equal(t, 0, m.RoundingMethod)
	assert.Equal(t, 0, m.RoundingDigit)
}

func TestOrderModel_Conversion(t *testing.T) {
	now := time.Now()
	paidAt := now.Add(-1 * time.Hour)
	order := &groupbuy.Order{
		ID:            "o1",
		GroupBuyID:    "gb1",
		UserID:        "u1",
		TotalAmount:   5000,
		PaymentStatus: groupbuy.PaymentStatusConfirmed,
		ContactInfo:   "John",
		ShippingAddress: "123 Main St",
		PaymentInfo: &groupbuy.PaymentInfo{
			Method:       "transfer",
			AccountLast5: "12345",
			PaidAt:       &paidAt,
			Amount:       5000,
		},
		Items: []*groupbuy.OrderItem{
			{ID: "oi1", OrderID: "o1", ProductID: "p1", SpecID: "s1", Quantity: 2, Status: groupbuy.OrderItemStatusOrdered, ProductName: "Product 1", SpecName: "Red", Price: 2500},
		},
		ShippingMethodID: "ship1",
		ShippingFee:      100,
		Note:             "Please hurry",
		CreatedAt:        now,
	}

	m := FromDomainOrder(order)
	assert.Equal(t, order.ID, m.ID)
	assert.Equal(t, order.GroupBuyID, m.GroupBuyID)
	assert.Equal(t, order.UserID, m.UserID)
	assert.Equal(t, int(order.PaymentStatus), m.PaymentStatus)
	assert.Equal(t, "transfer", m.PaymentMethod)
	assert.Equal(t, "12345", m.PaymentAccountLast5)
	assert.Equal(t, int64(5000), m.PaymentAmount)
	assert.NotNil(t, m.PaidAt)
	assert.Len(t, m.Items, 1)
	assert.Equal(t, 2, m.Items[0].Quantity)

	converted := m.ToDomain()
	assert.Equal(t, order.ID, converted.ID)
	assert.Equal(t, order.GroupBuyID, converted.GroupBuyID)
	assert.Equal(t, order.PaymentStatus, converted.PaymentStatus)
	assert.Equal(t, "transfer", converted.PaymentInfo.Method)
	assert.Equal(t, "12345", converted.PaymentInfo.AccountLast5)
	assert.Equal(t, int64(5000), converted.PaymentInfo.Amount)
	assert.Len(t, converted.Items, 1)
	assert.Equal(t, groupbuy.OrderItemStatusOrdered, converted.Items[0].Status)
}

func TestOrderModel_NilPaymentInfo(t *testing.T) {
	order := &groupbuy.Order{
		ID:        "o2",
		GroupBuyID: "gb1",
		UserID:    "u1",
	}
	m := FromDomainOrder(order)
	assert.Empty(t, m.PaymentMethod)
	assert.Empty(t, m.PaymentAccountLast5)
	assert.Nil(t, m.PaidAt)
	assert.Equal(t, int64(0), m.PaymentAmount)
}

func TestEventModel_Conversion(t *testing.T) {
	now := time.Now()
	start := now.Add(1 * time.Hour)
	end := now.Add(24 * time.Hour)
	deadline := now.Add(12 * time.Hour)
	itemStart := now.Add(2 * time.Hour)
	itemEnd := now.Add(3 * time.Hour)

	ev := &event.Event{
		ID:                   "e1",
		Title:                "Test Event",
		Description:          "A test event",
		CoverImage:           "http://cover",
		Status:               event.EventStatusActive,
		StartTime:            start,
		EndTime:              end,
		RegistrationDeadline: deadline,
		Location:             "Taipei",
		CreatorID:            "c1",
		ManagerIDs:           []string{"m1"},
		PaymentMethods:       []string{"cash", "transfer"},
		Items: []*event.EventItem{
			{ID: "ei1", EventID: "e1", Name: "Item 1", Price: 100, MinParticipants: 1, MaxParticipants: 10, StartTime: &itemStart, EndTime: &itemEnd, AllowMultiple: true},
		},
		Discounts: []*event.DiscountRule{
			{MinQuantity: 3, MinDistinctItems: 2, DiscountAmount: 50},
		},
		AllowException: true,
	}

	m := FromDomainEvent(ev)
	assert.Equal(t, ev.ID, m.ID)
	assert.Equal(t, ev.Title, m.Title)
	assert.Equal(t, int(ev.Status), m.Status)
	assert.Equal(t, ev.Location, m.Location)
	assert.Equal(t, ev.AllowException, m.AllowException)
	assert.Len(t, m.Managers, 1)
	assert.Len(t, m.Items, 1)
	assert.Len(t, m.Discounts, 1)
	assert.Equal(t, []string{"cash", "transfer"}, []string(m.PaymentMethods))

	converted := m.ToDomain()
	assert.Equal(t, ev.ID, converted.ID)
	assert.Equal(t, ev.Title, converted.Title)
	assert.Equal(t, ev.Status, converted.Status)
	assert.Equal(t, ev.Location, converted.Location)
	assert.Equal(t, ev.AllowException, converted.AllowException)
	assert.Len(t, converted.Items, 1)
	assert.Len(t, converted.Discounts, 1)
}

func TestDiscountRuleModel_Conversion(t *testing.T) {
	dr := &event.DiscountRule{
		MinQuantity:      5,
		MinDistinctItems: 3,
		DiscountAmount:   200,
	}

	m := FromDomainDiscountRule(dr)
	assert.NotEmpty(t, m.ID) // UUID generated
	assert.Equal(t, dr.MinQuantity, m.MinQuantity)
	assert.Equal(t, dr.MinDistinctItems, m.MinDistinctItems)
	assert.Equal(t, dr.DiscountAmount, m.DiscountAmount)

	converted := m.ToDomain()
	assert.Equal(t, dr.MinQuantity, converted.MinQuantity)
	assert.Equal(t, dr.DiscountAmount, converted.DiscountAmount)
}

func TestEventItemModel_Conversion(t *testing.T) {
	start := time.Now()
	end := start.Add(2 * time.Hour)
	item := &event.EventItem{
		ID:              "ei1",
		EventID:         "e1",
		Name:            "Workshop",
		Price:           500,
		MinParticipants: 5,
		MaxParticipants: 20,
		StartTime:       &start,
		EndTime:         &end,
		AllowMultiple:   true,
	}

	m := FromDomainEventItem(item)
	assert.Equal(t, item.ID, m.ID)
	assert.Equal(t, item.EventID, m.EventID)
	assert.Equal(t, item.Name, m.Name)
	assert.Equal(t, item.Price, m.Price)
	assert.Equal(t, item.AllowMultiple, m.AllowMultiple)

	converted := m.ToDomain()
	assert.Equal(t, item.ID, converted.ID)
	assert.Equal(t, item.Name, converted.Name)
	assert.Equal(t, item.Price, converted.Price)
	assert.Equal(t, item.AllowMultiple, converted.AllowMultiple)
	assert.NotNil(t, converted.StartTime)
	assert.NotNil(t, converted.EndTime)
}

func TestRegistrationModel_Conversion(t *testing.T) {
	reg := &event.Registration{
		ID:            "r1",
		EventID:       "e1",
		UserID:        "u1",
		Status:        event.RegistrationStatusConfirmed,
		PaymentStatus: event.PaymentStatusPaid,
		ContactInfo:   "John Doe",
		Notes:         "Vegetarian",
		TotalAmount:   1000,
		DiscountApplied: 100,
		SelectedItems: []*event.RegistrationItem{
			{EventItemID: "ei1", Quantity: 2},
		},
	}

	m := FromDomainRegistration(reg)
	assert.Equal(t, reg.ID, m.ID)
	assert.Equal(t, int(reg.Status), m.Status)
	assert.Equal(t, int(reg.PaymentStatus), m.PaymentStatus)
	assert.Len(t, m.SelectedItems, 1)
	assert.NotEmpty(t, m.SelectedItems[0].ID) // UUID generated

	converted := m.ToDomain()
	assert.Equal(t, reg.ID, converted.ID)
	assert.Equal(t, reg.Status, converted.Status)
	assert.Equal(t, reg.PaymentStatus, converted.PaymentStatus)
	assert.Equal(t, reg.ContactInfo, converted.ContactInfo)
	assert.Equal(t, reg.TotalAmount, converted.TotalAmount)
	assert.Equal(t, reg.DiscountApplied, converted.DiscountApplied)
	assert.Len(t, converted.SelectedItems, 1)
	assert.Equal(t, "ei1", converted.SelectedItems[0].EventItemID)
}

func TestRegistrationModel_WithUser(t *testing.T) {
	now := time.Now()
	reg := &Registration{
		ID:      "r1",
		EventID: "e1",
		UserID:  "u1",
		Status:  int(event.RegistrationStatusPending),
		User:    &User{ID: "u1", Name: "Alice", Email: "alice@test.com", CreatedAt: now, UpdatedAt: now},
	}
	converted := reg.ToDomain()
	assert.NotNil(t, converted.User)
	assert.Equal(t, "Alice", converted.User.Name)
}

func TestShippingConfigModel_Conversion(t *testing.T) {
	sc := &groupbuy.ShippingConfig{
		ID:    "sc1",
		Name:  "Home Delivery",
		Type:  groupbuy.ShippingTypeDelivery,
		Price: 150,
	}

	m := FromDomainShippingConfig(sc)
	assert.Equal(t, sc.ID, m.ID)
	assert.Equal(t, sc.Name, m.Name)
	assert.Equal(t, sc.Type, m.Type)
	assert.Equal(t, sc.Price, m.Price)

	converted := m.ToDomain()
	assert.Equal(t, sc.ID, converted.ID)
	assert.Equal(t, sc.Type, converted.Type)
}

func TestUserModel_ToDomainValid_Nil(t *testing.T) {
	var u *User
	assert.Nil(t, u.ToDomainValid())
}
