package handler

import (
	"math"

	v1 "github.com/hatsubosi/buygo-api/api/v1"
	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
)

func safeIntToInt32(n int) int32 {
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	if n < math.MinInt32 {
		return math.MinInt32
	}
	return int32(n)
}

func toProtoGroupBuyStatus(status groupbuy.GroupBuyStatus) v1.GroupBuyStatus {
	switch status {
	case groupbuy.GroupBuyStatusDraft:
		return v1.GroupBuyStatus_GROUP_BUY_STATUS_DRAFT
	case groupbuy.GroupBuyStatusActive:
		return v1.GroupBuyStatus_GROUP_BUY_STATUS_ACTIVE
	case groupbuy.GroupBuyStatusEnded:
		return v1.GroupBuyStatus_GROUP_BUY_STATUS_ENDED
	case groupbuy.GroupBuyStatusArchived:
		return v1.GroupBuyStatus_GROUP_BUY_STATUS_ARCHIVED
	default:
		return v1.GroupBuyStatus_GROUP_BUY_STATUS_UNSPECIFIED
	}
}

func toProtoRoundingMethod(method groupbuy.RoundingMethod) v1.RoundingMethod {
	switch method {
	case groupbuy.RoundingMethodFloor:
		return v1.RoundingMethod_ROUNDING_METHOD_FLOOR
	case groupbuy.RoundingMethodCeil:
		return v1.RoundingMethod_ROUNDING_METHOD_CEIL
	case groupbuy.RoundingMethodRound:
		return v1.RoundingMethod_ROUNDING_METHOD_ROUND
	default:
		return v1.RoundingMethod_ROUNDING_METHOD_UNSPECIFIED
	}
}

func toProtoShippingType(shippingType groupbuy.ShippingType) v1.ShippingType {
	switch shippingType {
	case groupbuy.ShippingTypeDelivery:
		return v1.ShippingType_SHIPPING_TYPE_DELIVERY
	case groupbuy.ShippingTypeStorePickup:
		return v1.ShippingType_SHIPPING_TYPE_STORE_PICKUP
	case groupbuy.ShippingTypeMeetup:
		return v1.ShippingType_SHIPPING_TYPE_MEETUP
	default:
		return v1.ShippingType_SHIPPING_TYPE_UNSPECIFIED
	}
}

func toProtoOrderItemStatus(status groupbuy.OrderItemStatus) v1.OrderItemStatus {
	switch status {
	case groupbuy.OrderItemStatusUnordered:
		return v1.OrderItemStatus_ITEM_STATUS_UNORDERED
	case groupbuy.OrderItemStatusOrdered:
		return v1.OrderItemStatus_ITEM_STATUS_ORDERED
	case groupbuy.OrderItemStatusArrivedOverseas:
		return v1.OrderItemStatus_ITEM_STATUS_ARRIVED_OVERSEAS
	case groupbuy.OrderItemStatusArrivedDomestic:
		return v1.OrderItemStatus_ITEM_STATUS_ARRIVED_DOMESTIC
	case groupbuy.OrderItemStatusReadyForPickup:
		return v1.OrderItemStatus_ITEM_STATUS_READY_FOR_PICKUP
	case groupbuy.OrderItemStatusSent:
		return v1.OrderItemStatus_ITEM_STATUS_SENT
	case groupbuy.OrderItemStatusFailed:
		return v1.OrderItemStatus_ITEM_STATUS_FAILED
	default:
		return v1.OrderItemStatus_ITEM_STATUS_UNSPECIFIED
	}
}

func toProtoGroupBuyPaymentStatus(status groupbuy.PaymentStatus) v1.PaymentStatus {
	switch status {
	case groupbuy.PaymentStatusUnset:
		return v1.PaymentStatus_PAYMENT_STATUS_UNSET
	case groupbuy.PaymentStatusSubmitted:
		return v1.PaymentStatus_PAYMENT_STATUS_SUBMITTED
	case groupbuy.PaymentStatusConfirmed:
		return v1.PaymentStatus_PAYMENT_STATUS_CONFIRMED
	case groupbuy.PaymentStatusRejected:
		return v1.PaymentStatus_PAYMENT_STATUS_REJECTED
	default:
		return v1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}

func toProtoEventStatus(status event.EventStatus) v1.EventStatus {
	switch status {
	case event.EventStatusDraft:
		return v1.EventStatus_EVENT_STATUS_DRAFT
	case event.EventStatusActive:
		return v1.EventStatus_EVENT_STATUS_ACTIVE
	case event.EventStatusEnded:
		return v1.EventStatus_EVENT_STATUS_ENDED
	case event.EventStatusArchived:
		return v1.EventStatus_EVENT_STATUS_ARCHIVED
	default:
		return v1.EventStatus_EVENT_STATUS_UNSPECIFIED
	}
}

func toProtoRegistrationStatus(status event.RegistrationStatus) v1.RegistrationStatus {
	switch status {
	case event.RegistrationStatusPending:
		return v1.RegistrationStatus_REGISTRATION_STATUS_PENDING
	case event.RegistrationStatusConfirmed:
		return v1.RegistrationStatus_REGISTRATION_STATUS_CONFIRMED
	case event.RegistrationStatusCancelled:
		return v1.RegistrationStatus_REGISTRATION_STATUS_CANCELLED
	default:
		return v1.RegistrationStatus_REGISTRATION_STATUS_UNSPECIFIED
	}
}

func toProtoEventPaymentStatus(status event.PaymentStatus) v1.PaymentStatus {
	switch status {
	case event.PaymentStatusUnpaid:
		return v1.PaymentStatus_PAYMENT_STATUS_UNSET
	case event.PaymentStatusSubmitted:
		return v1.PaymentStatus_PAYMENT_STATUS_SUBMITTED
	case event.PaymentStatusPaid:
		return v1.PaymentStatus_PAYMENT_STATUS_CONFIRMED
	case event.PaymentStatusRefunded:
		return v1.PaymentStatus_PAYMENT_STATUS_REJECTED
	default:
		return v1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}
