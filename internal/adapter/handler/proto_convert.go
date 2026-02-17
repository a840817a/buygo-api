package handler

import (
	"math"
	"time"

	v1 "github.com/hatsubosi/buygo-api/api/v1"
	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func toProtoEventItem(i *event.EventItem) *v1.EventItem {
	if i == nil {
		return nil
	}
	var start, end *timestamppb.Timestamp
	if i.StartTime != nil {
		start = timestamppb.New(*i.StartTime)
	}
	if i.EndTime != nil {
		end = timestamppb.New(*i.EndTime)
	}

	return &v1.EventItem{
		Id:              i.ID,
		Name:            i.Name,
		Price:           i.Price,
		MinParticipants: i.MinParticipants,
		MaxParticipants: i.MaxParticipants,
		StartTime:       start,
		EndTime:         end,
		AllowMultiple:   i.AllowMultiple,
	}
}

func toProtoDiscounts(rules []*event.DiscountRule) []*v1.DiscountRule {
	var res []*v1.DiscountRule
	for _, r := range rules {
		res = append(res, &v1.DiscountRule{
			MinQuantity:      int32(r.MinQuantity),
			MinDistinctItems: int32(r.MinDistinctItems),
			DiscountAmount:   int64(r.DiscountAmount),
		})
	}
	return res
}

func toProtoRegistration(r *event.Registration) *v1.Registration {
	if r == nil {
		return nil
	}

	// Flatten items for proto if needed or keep structure
	// Proto expects RegisterItem which matches logic
	var items []*v1.RegisterItem
	for _, i := range r.SelectedItems {
		items = append(items, &v1.RegisterItem{
			EventItemId: i.EventItemID,
			Quantity:    safeIntToInt32(i.Quantity),
		})
	}

	return &v1.Registration{
		Id:              r.ID,
		EventId:         r.EventID,
		UserId:          r.UserID,
		Status:          toProtoRegistrationStatus(r.Status),
		PaymentStatus:   toProtoEventPaymentStatus(r.PaymentStatus),
		ContactInfo:     r.ContactInfo,
		Notes:           r.Notes,
		TotalAmount:     r.TotalAmount,
		DiscountApplied: r.DiscountApplied,
		SelectedItems:   items,
		User:            toProtoUser(r.User),
	}
}

func toProtoUser(u *user.User) *v1.User {
	if u == nil {
		return nil
	}
	var role v1.UserRole
	switch u.Role {
	case user.UserRoleUser:
		role = v1.UserRole_USER_ROLE_USER
	case user.UserRoleCreator:
		role = v1.UserRole_USER_ROLE_CREATOR
	case user.UserRoleSysAdmin:
		role = v1.UserRole_USER_ROLE_SYS_ADMIN
	default:
		role = v1.UserRole_USER_ROLE_UNSPECIFIED
	}

	return &v1.User{
		Id:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		PhotoUrl: u.PhotoURL,
		Role:     role,
	}
}

func toTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}
