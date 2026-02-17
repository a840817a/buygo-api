package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

type GroupBuyService struct {
	repo groupbuy.Repository
}

func NewGroupBuyService(repo groupbuy.Repository) *GroupBuyService {
	return &GroupBuyService{
		repo: repo,
	}
}

// CreateGroupBuy: Only UserRoleCreator or Admin
func (s *GroupBuyService) CreateGroupBuy(ctx context.Context, title string, description string) (*groupbuy.GroupBuy, error) {
	usrID, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	// Check Role
	if _, _, err := requireRole(ctx, user.UserRoleCreator); err != nil {
		return nil, err
	}

	now := time.Now()
	// Default rounding config
	rounding := &groupbuy.RoundingConfig{Method: groupbuy.RoundingMethodFloor, Digit: 0}

	gb := &groupbuy.GroupBuy{
		ID:           uuid.New().String(),
		Title:        title,
		Description:  description,
		Status:       groupbuy.GroupBuyStatusDraft,
		ExchangeRate: 0.23, // Default Rate? Or 0
		Rounding:     rounding,
		CreatorID:    usrID,
		ManagerIDs:   []string{usrID}, // Creator is default manager
		CreatedAt:    now,
	}

	if err := s.repo.Create(ctx, gb); err != nil {
		return nil, err
	}

	return gb, nil
}

// GetGroupBuy: Public Read
func (s *GroupBuyService) GetGroupBuy(ctx context.Context, id string) (*groupbuy.GroupBuy, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// TODO: Filter draft/inactive projects if user is anonymous?
	// For now, allow viewing all if they have link, or check status
	return p, nil
}

// ListGroupBuys: Public Only
func (s *GroupBuyService) ListGroupBuys(ctx context.Context, limit, offset int) ([]*groupbuy.GroupBuy, error) {
	if limit <= 0 {
		limit = 10
	}
	// Always Public View
	return s.repo.List(ctx, limit, offset, "", false, false)
}

// ListManagerGroupBuys: Authenticated Manager/Admin View
func (s *GroupBuyService) ListManagerGroupBuys(ctx context.Context, limit, offset int) ([]*groupbuy.GroupBuy, error) {
	if limit <= 0 {
		limit = 10
	}
	userID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}
	isSysAdmin := role == int(user.UserRoleSysAdmin)
	return s.repo.List(ctx, limit, offset, userID, isSysAdmin, true)
}

// UpdateGroupBuy: Manager Only
func (s *GroupBuyService) UpdateGroupBuy(ctx context.Context, id string, title, desc string, status groupbuy.GroupBuyStatus, products []*groupbuy.Product, coverImage string, deadline *time.Time, shippingConfigs []*groupbuy.ShippingConfig, managerIDs []string, exchangeRate float64, rounding *groupbuy.RoundingConfig, sourceCurrency string) (*groupbuy.GroupBuy, error) {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Permission Check
	// 1. SysAdmin -> OK
	// 2. Creator -> OK
	// 3. Manager -> OK
	if !canManage(role, usrID, p) && p.CreatorID != usrID {
		return nil, ErrPermissionDenied
	}

	// Update Fields
	if title != "" {
		p.Title = title
	}
	if desc != "" {
		p.Description = desc
	}
	if status != groupbuy.GroupBuyStatusUnspecified {
		p.Status = status
	}
	if coverImage != "" {
		p.CoverImage = coverImage
	}
	if deadline != nil {
		p.Deadline = deadline
	}
	if shippingConfigs != nil {
		for _, sc := range shippingConfigs {
			if sc.ID == "" {
				sc.ID = uuid.New().String()
			}
		}
		p.ShippingConfigs = shippingConfigs
	}

	// Update Managers (Security: Only Creator or SysAdmin can update managers)
	if managerIDs != nil {
		// Strict Check: Only Creator or SysAdmin
		// Existing managers cannot add new managers unless they are Creator/SysAdmin.
		if role == int(user.UserRoleSysAdmin) || p.CreatorID == usrID {
			// Validate IDs? (Ideally, but optional for now. Frontend filters.)
			// We trust the provided IDs exist for MVP.
			p.ManagerIDs = managerIDs
		}
	}

	if sourceCurrency != "" {
		p.SourceCurrency = sourceCurrency
	}
	if exchangeRate > 0 {
		p.ExchangeRate = exchangeRate
	}
	if rounding != nil {
		p.Rounding = rounding
	}

	// Update Products
	if len(products) > 0 {
		for _, item := range products {
			// Apply Defaults if Missing (For both new and existing products)
			if item.ExchangeRate == 0 && p.ExchangeRate > 0 {
				item.ExchangeRate = p.ExchangeRate
			}
			if item.Rounding == nil && p.Rounding != nil {
				// Clone? Or Shared Struct? Safe to share struct if read-only logic
				item.Rounding = p.Rounding
			}

			if item.ID == "" {
				item.ID = uuid.New().String()
			}

			// Always Recalculate Final Price to ensure consistency/updates
			item.PriceFinal = s.CalculateFinalPrice(item.PriceOriginal, item.ExchangeRate, item.Rounding)
			item.GroupBuyID = p.ID

			// Ensure specs have IDs
			for _, spec := range item.Specs {
				if spec.ID == "" {
					spec.ID = uuid.New().String()
					spec.ProductID = item.ID
				}
			}
		}
		p.Products = products
	} else {
		if products != nil {
			p.Products = products
		}
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *GroupBuyService) GetMyGroupBuyOrder(ctx context.Context, groupBuyID string) (*groupbuy.Order, error) {
	userId, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	orders, err := s.repo.ListOrders(ctx, groupBuyID, userId)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, nil // No order found
	}
	// Return the most recent one? Or just the one (assuming single active order logic or just first)
	// Usually for this use case, if multiple, user manually manages. But returning latest is safer.
	// ListOrders implementation in memory didn't sort, postgres likely also not explicitly sorted but "Find" usually returns default order.
	// Let's return the last one (assuming append order) or explicitly first found?
	// Let's just return orders[0] for now.
	return orders[0], nil
}

func (s *GroupBuyService) UpdateOrder(ctx context.Context, orderID string, items []*groupbuy.OrderItem, note string) (*groupbuy.Order, error) {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Verify Owner or Manager
	gb, err := s.repo.GetByID(ctx, order.GroupBuyID)
	if err != nil {
		return nil, err
	}

	if role != int(user.UserRoleSysAdmin) && order.UserID != usrID && !gb.IsManager(usrID) {
		return nil, ErrPermissionDenied
	}

	// Validate Status - Allow edit only if not paid/locked?
	// Or if Items are not ordered yet.
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

	// Update Items
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

		// Snapshot
		item.ID = uuid.New().String()
		item.ProductName = prod.Name
		item.SpecName = specName
		item.Price = prod.PriceFinal
		item.OrderID = "" // Will be set by caller or Repo

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

// CreateOrder: User Only
func (s *GroupBuyService) CreateOrder(ctx context.Context, groupBuyID string, items []*groupbuy.OrderItem, contactInfo, shippingAddr, shippingMethodID, note string) (*groupbuy.Order, error) {
	usrID, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	// Validate and Snapshot Items
	validItems, total, err := s.prepareOrderItems(ctx, groupBuyID, items)
	if err != nil {
		return nil, err
	}

	p, err := s.repo.GetByID(ctx, groupBuyID)
	if err != nil {
		return nil, err
	}

	// Calculate Shipping Fee
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
			// If not found, ignore? or error?
			// Let's error if ID provided but not found.
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

// ListGroupBuyOrders: Manager Only
func (s *GroupBuyService) ListGroupBuyOrders(ctx context.Context, groupBuyID string) ([]*groupbuy.Order, error) {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	p, err := s.repo.GetByID(ctx, groupBuyID)
	if err != nil {
		return nil, err
	}

	// Verify Manager Access
	if !canManage(role, usrID, p) {
		return nil, ErrPermissionDenied
	}

	// Fetch orders for this project
	// Note: repo.ListOrders filters by groupBuyID and userID (if strict)
	// Here we want ALL orders for the group buy, regardless of user
	return s.repo.ListOrders(ctx, groupBuyID, "")
}

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

// CancelOrder: Owner or Admin. Only if items not yet processed beyond UNORDERED.
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

	// Cannot cancel if payment is already confirmed
	if order.PaymentStatus == groupbuy.PaymentStatusConfirmed {
		return ErrPaymentConfirmed
	}

	// Cannot cancel if any items have been processed (ordered from supplier or beyond)
	for _, item := range order.Items {
		if item.Status > groupbuy.OrderItemStatusUnordered &&
			item.Status != groupbuy.OrderItemStatusFailed {
			return ErrItemsProcessed
		}
	}

	// Mark all items as failed (cancelled)
	for _, item := range order.Items {
		item.Status = groupbuy.OrderItemStatusFailed
	}

	// Reset payment status
	order.PaymentStatus = groupbuy.PaymentStatusRejected

	return s.repo.UpdateOrder(ctx, order)
}

func (s *GroupBuyService) AddProduct(ctx context.Context, groupBuyID string, name string, priceOriginal int64, exchangeRate float64, specs []string) (*groupbuy.Product, error) {
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

	// Default rounding config logic
	rounding := &groupbuy.RoundingConfig{Method: groupbuy.RoundingMethodFloor, Digit: 0}
	// Use Group Buy Defaults if not provided (Wait, params only have rate? AddProduct signature might need update or infer from project if 0)

	rate := exchangeRate
	if rate == 0 {
		rate = p.ExchangeRate
	}
	if p.Rounding != nil {
		rounding = p.Rounding
	}

	productID := uuid.New().String()
	var productSpecs []*groupbuy.ProductSpec
	for _, specName := range specs {
		if specName == "" {
			continue
		}
		productSpecs = append(productSpecs, &groupbuy.ProductSpec{
			ID:        uuid.New().String(),
			ProductID: productID,
			Name:      specName,
		})
	}

	// Calculate Final Price
	priceFinal := s.CalculateFinalPrice(priceOriginal, rate, rounding)

	prod := &groupbuy.Product{
		ID:            productID,
		GroupBuyID:    groupBuyID,
		Name:          name,
		PriceOriginal: priceOriginal,
		ExchangeRate:  rate,
		Rounding:      rounding,
		PriceFinal:    priceFinal,
		Specs:         productSpecs,
	}

	if err := s.repo.AddProduct(ctx, prod); err != nil {
		return nil, err
	}

	return prod, nil
}

func (s *GroupBuyService) DeleteProduct(ctx context.Context, groupBuyID, productID string) error {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return err
	}

	p, err := s.repo.GetByID(ctx, groupBuyID)
	if err != nil {
		return err
	}

	if !canManage(role, usrID, p) {
		return ErrPermissionDenied
	}

	return s.repo.DeleteProduct(ctx, groupBuyID, productID)
}

func (s *GroupBuyService) GetMyOrders(ctx context.Context) ([]*groupbuy.Order, error) {
	usrID, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListOrders(ctx, "", usrID)
}

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

	// State Transition Logic (Strict FIFO Chain)
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

	return int32(n), ids, nil
}

func (s *GroupBuyService) CreateCategory(ctx context.Context, name string, specNames []string) (*groupbuy.Category, error) {
	_, _, err := requireRole(ctx, user.UserRoleSysAdmin)
	if err != nil {
		return nil, err
	}

	c, err := groupbuy.NewCategory(name, specNames)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateCategory(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *GroupBuyService) ListCategories(ctx context.Context) ([]*groupbuy.Category, error) {
	// Public access allowed? Or Authenticated?
	// Add Product form is manager only (Auth required).
	// Let's require Auth at least.
	_, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.ListCategories(ctx)
}

func (s *GroupBuyService) CalculateFinalPrice(original int64, rate float64, rounding *groupbuy.RoundingConfig) int64 {
	if original == 0 || rate == 0 {
		return 0
	}

	val := float64(original) * rate

	method := groupbuy.RoundingMethodFloor
	digit := 0

	if rounding != nil {
		method = rounding.Method
		digit = rounding.Digit
	}

	pow := math.Pow(10, float64(digit))
	val = val / pow

	switch method {
	case groupbuy.RoundingMethodCeil:
		val = math.Ceil(val)
	case groupbuy.RoundingMethodRound:
		val = math.Round(val)
	default:
		val = math.Floor(val)
	}

	val = val * pow
	return int64(val)
}
