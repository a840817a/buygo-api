package handler

import (
	"testing"
	"time"

	"github.com/buygo/buygo-api/internal/domain/event"
	"github.com/buygo/buygo-api/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestToProtoEvent(t *testing.T) {
	now := time.Now()
	deadline := now.Add(24 * time.Hour)
	startTime := now.Add(-time.Hour)
	endTime := now.Add(time.Hour)

	e := &event.Event{
		ID:                   "evt-1",
		Title:                "Test Event",
		Description:          "Desc",
		CoverImage:           "http://img.jpg",
		Status:               event.EventStatusActive,
		StartTime:            now,
		EndTime:              endTime,
		RegistrationDeadline: deadline,
		Location:             "Tokyo",
		CreatorID:            "creator-1",
		Creator:              &user.User{ID: "creator-1", Name: "Creator"},
		Managers: []*user.User{
			{ID: "mgr-1", Name: "Manager 1"},
		},
		PaymentMethods: []string{"bank_transfer", "cash"},
		AllowException: true,
		Items: []*event.EventItem{
			{
				ID:              "item-1",
				Name:            "VIP",
				Price:           500,
				MinParticipants: 1,
				MaxParticipants: 100,
				StartTime:       &startTime,
				EndTime:         &endTime,
				AllowMultiple:   true,
			},
		},
		Discounts: []*event.DiscountRule{
			{MinQuantity: 5, MinDistinctItems: 2, DiscountAmount: 100},
		},
	}

	proto := toProtoEvent(e)
	require.NotNil(t, proto)
	assert.Equal(t, "evt-1", proto.Id)
	assert.Equal(t, "Test Event", proto.Title)
	assert.Equal(t, "Desc", proto.Description)
	assert.Equal(t, "http://img.jpg", proto.CoverImageUrl)
	assert.Equal(t, "Tokyo", proto.Location)
	assert.True(t, proto.AllowModification)
	assert.Equal(t, []string{"bank_transfer", "cash"}, proto.PaymentMethods)

	// Creator
	require.NotNil(t, proto.Creator)
	assert.Equal(t, "creator-1", proto.Creator.Id)

	// Managers
	require.Len(t, proto.Managers, 1)
	assert.Equal(t, "mgr-1", proto.Managers[0].Id)

	// Items
	require.Len(t, proto.Items, 1)
	assert.Equal(t, "item-1", proto.Items[0].Id)
	assert.Equal(t, "VIP", proto.Items[0].Name)
	assert.Equal(t, int64(500), proto.Items[0].Price)
	assert.True(t, proto.Items[0].AllowMultiple)

	// Discounts
	require.Len(t, proto.Discounts, 1)
	assert.Equal(t, int32(5), proto.Discounts[0].MinQuantity)
	assert.Equal(t, int64(100), proto.Discounts[0].DiscountAmount)
}

func TestToProtoEvent_Nil(t *testing.T) {
	assert.Nil(t, toProtoEvent(nil))
}

func TestToProtoEvent_Empty(t *testing.T) {
	e := &event.Event{}
	proto := toProtoEvent(e)
	require.NotNil(t, proto)
	assert.Empty(t, proto.Id)
	assert.Nil(t, proto.Creator)
	assert.Empty(t, proto.Managers)
	assert.Empty(t, proto.Items)
	assert.Empty(t, proto.Discounts)
}

func TestToProtoRegistration(t *testing.T) {
	reg := &event.Registration{
		ID:              "reg-1",
		EventID:         "evt-1",
		UserID:          "user-1",
		Status:          event.RegistrationStatusConfirmed,
		PaymentStatus:   2,
		ContactInfo:     "0912345678",
		Notes:           "notes here",
		TotalAmount:     500,
		DiscountApplied: 50,
		SelectedItems: []*event.RegistrationItem{
			{EventItemID: "item-1", Quantity: 3},
			{EventItemID: "item-2", Quantity: 1},
		},
		User: &user.User{ID: "user-1", Name: "User One"},
	}

	proto := toProtoRegistration(reg)
	require.NotNil(t, proto)
	assert.Equal(t, "reg-1", proto.Id)
	assert.Equal(t, "evt-1", proto.EventId)
	assert.Equal(t, "user-1", proto.UserId)
	assert.Equal(t, int64(500), proto.TotalAmount)
	assert.Equal(t, int64(50), proto.DiscountApplied)
	assert.Equal(t, "0912345678", proto.ContactInfo)
	assert.Equal(t, "notes here", proto.Notes)

	// Items
	require.Len(t, proto.SelectedItems, 2)
	assert.Equal(t, "item-1", proto.SelectedItems[0].EventItemId)
	assert.Equal(t, int32(3), proto.SelectedItems[0].Quantity)

	// User
	require.NotNil(t, proto.User)
	assert.Equal(t, "user-1", proto.User.Id)
}

func TestToProtoRegistration_Nil(t *testing.T) {
	assert.Nil(t, toProtoRegistration(nil))
}

func TestToProtoRegistration_NoItems(t *testing.T) {
	reg := &event.Registration{
		ID:            "reg-1",
		SelectedItems: nil,
	}
	proto := toProtoRegistration(reg)
	require.NotNil(t, proto)
	assert.Empty(t, proto.SelectedItems)
}

func TestToProtoEventItem(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Hour)

	item := &event.EventItem{
		ID:              "item-1",
		Name:            "Ticket",
		Price:           200,
		MinParticipants: 5,
		MaxParticipants: 50,
		StartTime:       &start,
		EndTime:         &end,
		AllowMultiple:   true,
	}

	proto := toProtoEventItem(item)
	require.NotNil(t, proto)
	assert.Equal(t, "item-1", proto.Id)
	assert.Equal(t, "Ticket", proto.Name)
	assert.Equal(t, int64(200), proto.Price)
	assert.Equal(t, int32(5), proto.MinParticipants)
	assert.Equal(t, int32(50), proto.MaxParticipants)
	assert.True(t, proto.AllowMultiple)
	assert.NotNil(t, proto.StartTime)
	assert.NotNil(t, proto.EndTime)
}

func TestToProtoEventItem_Nil(t *testing.T) {
	assert.Nil(t, toProtoEventItem(nil))
}

func TestToProtoEventItem_NoTimes(t *testing.T) {
	item := &event.EventItem{ID: "item-1", Name: "Basic"}
	proto := toProtoEventItem(item)
	require.NotNil(t, proto)
	assert.Nil(t, proto.StartTime)
	assert.Nil(t, proto.EndTime)
}

func TestToProtoDiscounts(t *testing.T) {
	rules := []*event.DiscountRule{
		{MinQuantity: 2, MinDistinctItems: 1, DiscountAmount: 20},
		{MinQuantity: 5, MinDistinctItems: 3, DiscountAmount: 80},
	}

	result := toProtoDiscounts(rules)
	require.Len(t, result, 2)
	assert.Equal(t, int32(2), result[0].MinQuantity)
	assert.Equal(t, int64(80), result[1].DiscountAmount)
}

func TestToProtoDiscounts_Nil(t *testing.T) {
	assert.Nil(t, toProtoDiscounts(nil))
}

func TestToTime(t *testing.T) {
	ts := timestamppb.Now()
	result := toTime(ts)
	require.NotNil(t, result)
	assert.WithinDuration(t, time.Now(), *result, 2*time.Second)
}

func TestToTime_Nil(t *testing.T) {
	result := toTime(nil)
	assert.Nil(t, result)
}
