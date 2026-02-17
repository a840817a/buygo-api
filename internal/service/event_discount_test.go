package service

import (
	"testing"

	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/stretchr/testify/assert"
)

func TestCalculateTotal_DistinctItems(t *testing.T) {
	svc := &EventService{}

	item1 := &event.EventItem{ID: "item1", Price: 100}
	item2 := &event.EventItem{ID: "item2", Price: 100}

	e := &event.Event{
		Items: []*event.EventItem{item1, item2},
		Discounts: []*event.DiscountRule{
			{
				MinQuantity:      2,
				MinDistinctItems: 2,
				DiscountAmount:   50,
			},
		},
	}

	// Case 1: 2 items, same type -> No Discount
	regItems1 := []*event.RegistrationItem{
		{EventItemID: "item1", Quantity: 2},
	}
	total1, discount1 := svc.calculateTotal(e, regItems1)
	assert.Equal(t, int64(200), total1)
	assert.Equal(t, int64(0), discount1)

	// Case 2: 2 items, different types -> Discount Applied
	regItems2 := []*event.RegistrationItem{
		{EventItemID: "item1", Quantity: 1},
		{EventItemID: "item2", Quantity: 1},
	}
	total2, discount2 := svc.calculateTotal(e, regItems2)
	assert.Equal(t, int64(150), total2)
	assert.Equal(t, int64(50), discount2)
}

func TestCalculateTotal_MultipleRules_PicksBest(t *testing.T) {
	svc := &EventService{}

	e := &event.Event{
		Items: []*event.EventItem{
			{ID: "a", Price: 100},
			{ID: "b", Price: 100},
		},
		Discounts: []*event.DiscountRule{
			{MinQuantity: 2, MinDistinctItems: 1, DiscountAmount: 20},
			{MinQuantity: 2, MinDistinctItems: 2, DiscountAmount: 50},
			{MinQuantity: 3, MinDistinctItems: 1, DiscountAmount: 80},
		},
	}

	// 2 distinct items, qty=2 → qualifies for rules 1 & 2, picks 50
	items := []*event.RegistrationItem{
		{EventItemID: "a", Quantity: 1},
		{EventItemID: "b", Quantity: 1},
	}
	total, discount := svc.calculateTotal(e, items)
	assert.Equal(t, int64(50), discount, "Should pick highest qualifying discount")
	assert.Equal(t, int64(150), total)
}

func TestCalculateTotal_DiscountCapped(t *testing.T) {
	svc := &EventService{}

	e := &event.Event{
		Items: []*event.EventItem{
			{ID: "a", Price: 30},
		},
		Discounts: []*event.DiscountRule{
			{MinQuantity: 1, MinDistinctItems: 1, DiscountAmount: 100}, // 100 > 30
		},
	}

	items := []*event.RegistrationItem{{EventItemID: "a", Quantity: 1}}
	total, discount := svc.calculateTotal(e, items)
	assert.Equal(t, int64(30), discount, "Discount should be capped at subtotal")
	assert.Equal(t, int64(0), total, "Total should be 0 after capped discount")
}

func TestCalculateTotal_ZeroItems(t *testing.T) {
	svc := &EventService{}

	e := &event.Event{
		Items: []*event.EventItem{{ID: "a", Price: 100}},
		Discounts: []*event.DiscountRule{
			{MinQuantity: 1, MinDistinctItems: 1, DiscountAmount: 50},
		},
	}

	total, discount := svc.calculateTotal(e, nil)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, int64(0), discount)
}

func TestCalculateTotal_NoDiscounts(t *testing.T) {
	svc := &EventService{}

	e := &event.Event{
		Items:     []*event.EventItem{{ID: "a", Price: 100}},
		Discounts: nil,
	}

	items := []*event.RegistrationItem{{EventItemID: "a", Quantity: 3}}
	total, discount := svc.calculateTotal(e, items)
	assert.Equal(t, int64(300), total)
	assert.Equal(t, int64(0), discount)
}
