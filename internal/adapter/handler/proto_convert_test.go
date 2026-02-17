package handler

import (
	"math"
	"testing"

	v1 "github.com/hatsubosi/buygo-api/api/v1"
	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
)

func TestSafeIntToInt32(t *testing.T) {
	if got := safeIntToInt32(math.MaxInt32 + 10); got != math.MaxInt32 {
		t.Fatalf("safeIntToInt32(max overflow) = %d, want %d", got, int32(math.MaxInt32))
	}
	if got := safeIntToInt32(math.MinInt32 - 10); got != math.MinInt32 {
		t.Fatalf("safeIntToInt32(min overflow) = %d, want %d", got, int32(math.MinInt32))
	}
	if got := safeIntToInt32(123); got != 123 {
		t.Fatalf("safeIntToInt32(123) = %d, want 123", got)
	}
}

func TestProtoEnumMappings_GroupBuy(t *testing.T) {
	statusCases := []struct {
		in   groupbuy.GroupBuyStatus
		want v1.GroupBuyStatus
	}{
		{groupbuy.GroupBuyStatusDraft, v1.GroupBuyStatus_GROUP_BUY_STATUS_DRAFT},
		{groupbuy.GroupBuyStatusActive, v1.GroupBuyStatus_GROUP_BUY_STATUS_ACTIVE},
		{groupbuy.GroupBuyStatusEnded, v1.GroupBuyStatus_GROUP_BUY_STATUS_ENDED},
		{groupbuy.GroupBuyStatusArchived, v1.GroupBuyStatus_GROUP_BUY_STATUS_ARCHIVED},
		{groupbuy.GroupBuyStatusUnspecified, v1.GroupBuyStatus_GROUP_BUY_STATUS_UNSPECIFIED},
	}
	for _, tc := range statusCases {
		if got := toProtoGroupBuyStatus(tc.in); got != tc.want {
			t.Fatalf("toProtoGroupBuyStatus(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}

	roundCases := []struct {
		in   groupbuy.RoundingMethod
		want v1.RoundingMethod
	}{
		{groupbuy.RoundingMethodFloor, v1.RoundingMethod_ROUNDING_METHOD_FLOOR},
		{groupbuy.RoundingMethodCeil, v1.RoundingMethod_ROUNDING_METHOD_CEIL},
		{groupbuy.RoundingMethodRound, v1.RoundingMethod_ROUNDING_METHOD_ROUND},
		{groupbuy.RoundingMethodUnspecified, v1.RoundingMethod_ROUNDING_METHOD_UNSPECIFIED},
	}
	for _, tc := range roundCases {
		if got := toProtoRoundingMethod(tc.in); got != tc.want {
			t.Fatalf("toProtoRoundingMethod(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}

	shippingCases := []struct {
		in   groupbuy.ShippingType
		want v1.ShippingType
	}{
		{groupbuy.ShippingTypeDelivery, v1.ShippingType_SHIPPING_TYPE_DELIVERY},
		{groupbuy.ShippingTypeStorePickup, v1.ShippingType_SHIPPING_TYPE_STORE_PICKUP},
		{groupbuy.ShippingTypeMeetup, v1.ShippingType_SHIPPING_TYPE_MEETUP},
		{groupbuy.ShippingTypeUnspecified, v1.ShippingType_SHIPPING_TYPE_UNSPECIFIED},
	}
	for _, tc := range shippingCases {
		if got := toProtoShippingType(tc.in); got != tc.want {
			t.Fatalf("toProtoShippingType(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}

	orderItemCases := []struct {
		in   groupbuy.OrderItemStatus
		want v1.OrderItemStatus
	}{
		{groupbuy.OrderItemStatusUnordered, v1.OrderItemStatus_ITEM_STATUS_UNORDERED},
		{groupbuy.OrderItemStatusOrdered, v1.OrderItemStatus_ITEM_STATUS_ORDERED},
		{groupbuy.OrderItemStatusArrivedOverseas, v1.OrderItemStatus_ITEM_STATUS_ARRIVED_OVERSEAS},
		{groupbuy.OrderItemStatusArrivedDomestic, v1.OrderItemStatus_ITEM_STATUS_ARRIVED_DOMESTIC},
		{groupbuy.OrderItemStatusReadyForPickup, v1.OrderItemStatus_ITEM_STATUS_READY_FOR_PICKUP},
		{groupbuy.OrderItemStatusSent, v1.OrderItemStatus_ITEM_STATUS_SENT},
		{groupbuy.OrderItemStatusFailed, v1.OrderItemStatus_ITEM_STATUS_FAILED},
		{groupbuy.OrderItemStatusUnspecified, v1.OrderItemStatus_ITEM_STATUS_UNSPECIFIED},
	}
	for _, tc := range orderItemCases {
		if got := toProtoOrderItemStatus(tc.in); got != tc.want {
			t.Fatalf("toProtoOrderItemStatus(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}

	paymentCases := []struct {
		in   groupbuy.PaymentStatus
		want v1.PaymentStatus
	}{
		{groupbuy.PaymentStatusUnset, v1.PaymentStatus_PAYMENT_STATUS_UNSET},
		{groupbuy.PaymentStatusSubmitted, v1.PaymentStatus_PAYMENT_STATUS_SUBMITTED},
		{groupbuy.PaymentStatusConfirmed, v1.PaymentStatus_PAYMENT_STATUS_CONFIRMED},
		{groupbuy.PaymentStatusRejected, v1.PaymentStatus_PAYMENT_STATUS_REJECTED},
		{groupbuy.PaymentStatusUnspecified, v1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED},
	}
	for _, tc := range paymentCases {
		if got := toProtoGroupBuyPaymentStatus(tc.in); got != tc.want {
			t.Fatalf("toProtoGroupBuyPaymentStatus(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestProtoEnumMappings_Event(t *testing.T) {
	eventStatusCases := []struct {
		in   event.EventStatus
		want v1.EventStatus
	}{
		{event.EventStatusDraft, v1.EventStatus_EVENT_STATUS_DRAFT},
		{event.EventStatusActive, v1.EventStatus_EVENT_STATUS_ACTIVE},
		{event.EventStatusEnded, v1.EventStatus_EVENT_STATUS_ENDED},
		{event.EventStatusArchived, v1.EventStatus_EVENT_STATUS_ARCHIVED},
		{event.EventStatusUnspecified, v1.EventStatus_EVENT_STATUS_UNSPECIFIED},
	}
	for _, tc := range eventStatusCases {
		if got := toProtoEventStatus(tc.in); got != tc.want {
			t.Fatalf("toProtoEventStatus(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}

	regStatusCases := []struct {
		in   event.RegistrationStatus
		want v1.RegistrationStatus
	}{
		{event.RegistrationStatusPending, v1.RegistrationStatus_REGISTRATION_STATUS_PENDING},
		{event.RegistrationStatusConfirmed, v1.RegistrationStatus_REGISTRATION_STATUS_CONFIRMED},
		{event.RegistrationStatusCancelled, v1.RegistrationStatus_REGISTRATION_STATUS_CANCELLED},
		{event.RegistrationStatusUnspecified, v1.RegistrationStatus_REGISTRATION_STATUS_UNSPECIFIED},
	}
	for _, tc := range regStatusCases {
		if got := toProtoRegistrationStatus(tc.in); got != tc.want {
			t.Fatalf("toProtoRegistrationStatus(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}

	eventPaymentCases := []struct {
		in   event.PaymentStatus
		want v1.PaymentStatus
	}{
		{event.PaymentStatusUnpaid, v1.PaymentStatus_PAYMENT_STATUS_UNSET},
		{event.PaymentStatusSubmitted, v1.PaymentStatus_PAYMENT_STATUS_SUBMITTED},
		{event.PaymentStatusPaid, v1.PaymentStatus_PAYMENT_STATUS_CONFIRMED},
		{event.PaymentStatusRefunded, v1.PaymentStatus_PAYMENT_STATUS_REJECTED},
		{event.PaymentStatusUnspecified, v1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED},
	}
	for _, tc := range eventPaymentCases {
		if got := toProtoEventPaymentStatus(tc.in); got != tc.want {
			t.Fatalf("toProtoEventPaymentStatus(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
