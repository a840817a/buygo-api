package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

// GetMyGroupBuyOrder returns the current user's order for a specific group buy.
func (s *GroupBuyService) GetMyGroupBuyOrder(ctx context.Context, groupBuyID string) (*groupbuy.Order, error) {
	userID, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	orders, err := s.repo.ListOrders(ctx, groupBuyID, userID)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, nil
	}
	return orders[0], nil
}

// CreateOrder creates a new order for the current user.
func (s *GroupBuyService) CreateOrder(ctx context.Context, groupBuyID string, items []*groupbuy.OrderItem, contactInfo, shippingAddr, shippingMethodID, note string) (*groupbuy.Order, error) {
	usrID, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	validItems, total, err := s.prepareOrderItems(ctx, groupBuyID, items)
	if err != nil {
		return nil, err
	}

	p, err := s.repo.GetByID(ctx, groupBuyID)
	if err != nil {
		return nil, err
	}

	var shippingFee int64
	if shippingMethodID != "" {
		found := false
		for _, sc := range p.ShippingConfigs {
			if sc.ID == shippingMethodID {
				shippingFee = sc.Price
				found = true
				break
			}
		}
		if !found {
			return nil, ErrInvalidShippingMethod
		}
	}

	order := &groupbuy.Order{
		ID:               uuid.New().String(),
		GroupBuyID:       groupBuyID,
		UserID:           usrID,
		Items:            validItems,
		TotalAmount:      total + shippingFee,
		CreatedAt:        time.Now(),
		PaymentStatus:    groupbuy.PaymentStatusUnset,
		ContactInfo:      contactInfo,
		ShippingAddress:  shippingAddr,
		ShippingMethodID: shippingMethodID,
		ShippingFee:      shippingFee,
		Note:             note,
	}

	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

// UpdateOrder updates an existing order's items and/or note.
func (s *GroupBuyService) UpdateOrder(ctx context.Context, orderID string, items []*groupbuy.OrderItem, note string) (*groupbuy.Order, error) {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	gb, err := s.repo.GetByID(ctx, order.GroupBuyID)
	if err != nil {
		return nil, err
	}

	if role != int(user.UserRoleSysAdmin) && order.UserID != usrID && !gb.IsManager(usrID) {
		return nil, ErrPermissionDenied
	}

	if order.PaymentStatus == groupbuy.PaymentStatusConfirmed {
		return nil, ErrPaymentConfirmed
	}

	isMgr := gb.IsManager(usrID)
	if !isMgr {
		for _, i := range order.Items {
			if i.Status > groupbuy.OrderItemStatusUnordered {
				return nil, ErrItemsProcessed
			}
		}
	}

	validItems, total, err := s.prepareOrderItems(ctx, order.GroupBuyID, items)
	if err != nil {
		return nil, err
	}

	if !isMgr {
		for _, item := range validItems {
			item.Status = groupbuy.OrderItemStatusUnordered
		}
	}

	order.Items = validItems
	order.TotalAmount = total
	if note != "" {
		order.Note = note
	}

	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

// prepareOrderItems validates and snapshots order items against the group buy's current products.
func (s *GroupBuyService) prepareOrderItems(ctx context.Context, groupBuyID string, inputItems []*groupbuy.OrderItem) ([]*groupbuy.OrderItem, int64, error) {
	gb, err := s.repo.GetByID(ctx, groupBuyID)
	if err != nil {
		return nil, 0, err
	}

	if gb.Status != groupbuy.GroupBuyStatusActive {
		return nil, 0, ErrNotActive
	}

	productMap := make(map[string]*groupbuy.Product)
	for _, prod := range gb.Products {
		productMap[prod.ID] = prod
	}

	var total int64
	var validItems []*groupbuy.OrderItem

	for _, item := range inputItems {
		prod, ok := productMap[item.ProductID]
		if !ok {
			return nil, 0, fmt.Errorf("%s: %w", item.ProductID, ErrProductNotFound)
		}

		specName := "Default"
		if item.SpecID != "" {
			foundSpec := false
			for _, sp := range prod.Specs {
				if sp.ID == item.SpecID {
					specName = sp.Name
					foundSpec = true
					break
				}
			}
			if !foundSpec {
				return nil, 0, fmt.Errorf("%s: %w", item.SpecID, ErrSpecNotFound)
			}
		}

		item.ID = uuid.New().String()
		item.ProductName = prod.Name
		item.SpecName = specName
		item.Price = prod.PriceFinal
		item.OrderID = ""

		if item.Status == groupbuy.OrderItemStatusUnspecified {
			item.Status = groupbuy.OrderItemStatusUnordered
		}
		if item.Quantity <= 0 {
			return nil, 0, ErrInvalidQuantity
		}

		total += item.Price * int64(item.Quantity)
		validItems = append(validItems, item)
	}

	return validItems, total, nil
}

// UpdatePaymentInfo updates payment info and/or shipping/contact details on an order.
func (s *GroupBuyService) UpdatePaymentInfo(ctx context.Context, orderID string, method, account string, contact, shipping string, paidAt *time.Time, amount int64) (*groupbuy.Order, error) {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	p, err := s.repo.GetByID(ctx, order.GroupBuyID)
	if err != nil {
		return nil, err
	}

	if role != int(user.UserRoleSysAdmin) && order.UserID != usrID && !p.IsManager(usrID) {
		return nil, ErrPermissionDenied
	}

	if order.PaymentStatus == groupbuy.PaymentStatusConfirmed {
		return nil, ErrPaymentConfirmed
	}

	updated := false
	if method != "" || account != "" || paidAt != nil || amount != 0 {
		if order.PaymentInfo == nil {
			order.PaymentInfo = &groupbuy.PaymentInfo{}
		}
		if method != "" {
			order.PaymentInfo.Method = method
		}
		if account != "" {
			order.PaymentInfo.AccountLast5 = account
		}
		if paidAt != nil {
			order.PaymentInfo.PaidAt = paidAt
		}
		if amount != 0 {
			order.PaymentInfo.Amount = amount
		}
		if order.PaymentInfo.Method != "" && (order.PaymentInfo.AccountLast5 != "" || order.PaymentInfo.Method == "Cash") {
			order.PaymentStatus = groupbuy.PaymentStatusSubmitted
		}
		updated = true
	}

	if contact != "" {
		order.ContactInfo = contact
		updated = true
	}
	if shipping != "" {
		order.ShippingAddress = shipping
		updated = true
	}

	if updated {
		if err := s.repo.UpdateOrder(ctx, order); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// ListGroupBuyOrders returns all orders for a group buy (manager only).
func (s *GroupBuyService) ListGroupBuyOrders(ctx context.Context, groupBuyID string) ([]*groupbuy.Order, error) {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	p, err := s.repo.GetByID(ctx, groupBuyID)
	if err != nil {
		return nil, err
	}

	if !canManage(role, usrID, p) {
		return nil, ErrPermissionDenied
	}

	return s.repo.ListOrders(ctx, groupBuyID, "")
}

// ConfirmPayment updates the payment status of an order (manager only).
func (s *GroupBuyService) ConfirmPayment(ctx context.Context, orderID string, status groupbuy.PaymentStatus) error {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return err
	}

	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	p, err := s.repo.GetByID(ctx, order.GroupBuyID)
	if err != nil {
		return err
	}

	if !canManage(role, usrID, p) {
		return ErrPermissionDenied
	}

	return s.repo.UpdateOrderPaymentStatus(ctx, orderID, status)
}

// CancelOrder cancels an order (owner or admin only).
func (s *GroupBuyService) CancelOrder(ctx context.Context, orderID string) error {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return err
	}

	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	if role != int(user.UserRoleSysAdmin) && order.UserID != usrID {
		return ErrPermissionDenied
	}

	if order.PaymentStatus == groupbuy.PaymentStatusConfirmed {
		return ErrPaymentConfirmed
	}

	for _, item := range order.Items {
		if item.Status > groupbuy.OrderItemStatusUnordered &&
			item.Status != groupbuy.OrderItemStatusFailed {
			return ErrItemsProcessed
		}
	}

	for _, item := range order.Items {
		item.Status = groupbuy.OrderItemStatusFailed
	}
	order.PaymentStatus = groupbuy.PaymentStatusRejected

	return s.repo.UpdateOrder(ctx, order)
}

// GetMyOrders returns all orders for the current user.
func (s *GroupBuyService) GetMyOrders(ctx context.Context) ([]*groupbuy.Order, error) {
	usrID, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.ListOrders(ctx, "", usrID)
}

// BatchUpdateStatus advances a batch of order items to the next status (manager only).
func (s *GroupBuyService) BatchUpdateStatus(ctx context.Context, groupBuyID string, specID string, targetStatus int, count int32) (int32, []string, error) {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return 0, nil, err
	}

	p, err := s.repo.GetByID(ctx, groupBuyID)
	if err != nil {
		return 0, nil, err
	}

	if !canManage(role, usrID, p) {
		return 0, nil, ErrPermissionDenied
	}

	var fromStatuses []int
	switch groupbuy.OrderItemStatus(targetStatus) {
	case groupbuy.OrderItemStatusOrdered:
		fromStatuses = []int{int(groupbuy.OrderItemStatusUnordered), int(groupbuy.OrderItemStatusUnspecified)}
	case groupbuy.OrderItemStatusArrivedOverseas:
		fromStatuses = []int{int(groupbuy.OrderItemStatusOrdered)}
	case groupbuy.OrderItemStatusArrivedDomestic:
		fromStatuses = []int{int(groupbuy.OrderItemStatusArrivedOverseas)}
	case groupbuy.OrderItemStatusReadyForPickup:
		fromStatuses = []int{int(groupbuy.OrderItemStatusArrivedDomestic)}
	case groupbuy.OrderItemStatusSent:
		fromStatuses = []int{int(groupbuy.OrderItemStatusReadyForPickup)}
	default:
		return 0, nil, ErrInvalidStatus
	}

	if count <= 0 {
		return 0, nil, nil
	}

	n, ids, err := s.repo.BatchUpdateOrderItemStatus(ctx, groupBuyID, specID, fromStatuses, targetStatus, int(count))
	if err != nil {
		return 0, nil, err
	}
	if n > (1<<31-1) || n < -(1<<31) {
		return 0, nil, ErrInvalidQuantity
	}
	return int32(n), ids, nil
}
