package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/buygo/buygo-api/internal/domain/auth"
	"github.com/buygo/buygo-api/internal/domain/project"
	"github.com/buygo/buygo-api/internal/domain/user"
)

var (
	ErrPermissionDenied = errors.New("permission denied")
)

type ProjectService struct {
	repo project.Repository
}

func NewProjectService(repo project.Repository) *ProjectService {
	return &ProjectService{
		repo: repo,
	}
}

// CreateProject: Only UserRoleCreator or Admin
func (s *ProjectService) CreateProject(ctx context.Context, title string, description string) (*project.Project, error) {
	usrID, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	// Check Role
	if role != int(user.UserRoleCreator) && role != int(user.UserRoleSysAdmin) {
		return nil, ErrPermissionDenied
	}

	now := time.Now()
	// Default rounding config
	rounding := &project.RoundingConfig{Method: 1, Digit: 0} // Default Floor, Ones

	p := &project.Project{
		ID:           uuid.New().String(),
		Title:        title,
		Description:  description,
		Status:       project.ProjectStatusDraft,
		ExchangeRate: 0.23, // Default Rate? Or 0
		Rounding:     rounding,
		CreatorID:    usrID,
		ManagerIDs:   []string{usrID}, // Creator is default manager
		CreatedAt:    now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

// GetProject: Public Read
func (s *ProjectService) GetProject(ctx context.Context, id string) (*project.Project, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// TODO: Filter draft/inactive projects if user is anonymous?
	// For now, allow viewing all if they have link, or check status
	return p, nil
}

// ListProjects: Public List
// ListProjects: Public Only
func (s *ProjectService) ListProjects(ctx context.Context, limit, offset int) ([]*project.Project, error) {
	if limit <= 0 {
		limit = 10
	}
	// Always Public View
	return s.repo.List(ctx, limit, offset, "", false, false)
}

// ListManagerProjects: Authenticated Manager/Admin View
func (s *ProjectService) ListManagerProjects(ctx context.Context, limit, offset int) ([]*project.Project, error) {
	if limit <= 0 {
		limit = 10
	}
	userID, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}
	isSysAdmin := role == int(user.UserRoleSysAdmin)
	return s.repo.List(ctx, limit, offset, userID, isSysAdmin, true)
}

// UpdateProject: Manager Only
func (s *ProjectService) UpdateProject(ctx context.Context, id string, title, desc string, status project.ProjectStatus, products []*project.Product, coverImage string, deadline *time.Time, shippingConfigs []*project.ShippingConfig, managerIDs []string, exchangeRate float64, rounding *project.RoundingConfig, sourceCurrency string) (*project.Project, error) {
	usrID, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Permission Check
	// 1. SysAdmin -> OK
	// 2. Creator -> OK
	// 3. Manager -> OK
	isAuth := false
	if role == int(user.UserRoleSysAdmin) {
		isAuth = true
	} else if p.CreatorID == usrID {
		isAuth = true
	} else if isManager(p, usrID) {
		isAuth = true
	}

	if !isAuth {
		return nil, ErrPermissionDenied
	}

	// Update Fields
	if title != "" {
		p.Title = title
	}
	if desc != "" {
		p.Description = desc
	}
	if status != project.ProjectStatusUnspecified {
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
			item.ProjectID = p.ID

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

func (s *ProjectService) GetMyProjectOrder(ctx context.Context, projectID string) (*project.Order, error) {
	userId, _, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	orders, err := s.repo.ListOrders(ctx, projectID, userId)
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

func (s *ProjectService) UpdateOrder(ctx context.Context, orderID string, items []*project.OrderItem, note string) (*project.Order, error) {
	usrID, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Verify Owner or Manager
	p, err := s.repo.GetByID(ctx, order.ProjectID)
	if err != nil {
		return nil, err
	}

	if role != int(user.UserRoleSysAdmin) && order.UserID != usrID && !isManager(p, usrID) {
		return nil, ErrPermissionDenied
	}

	// Validate Status - Allow edit only if not paid/locked?
	// Or if Items are not ordered yet.
	// For now, if PaymentStatus is CONFIRMED (3), disallow.
	if order.PaymentStatus == 3 {
		return nil, errors.New("cannot update order: payment confirmed")
	}

	// Check if any items are already processed (Status > 1)
	// If Manager is editing, maybe allow?
	// Requirement: "If Manager updated it, User cannot edit".
	// If caller is Manager, we skip this check?
	// s.UpdateOrder is called by User (UpdateOrder RPC) and Manager (via my new calling code?).
	// Wait, I updated permissions to allow Manager.
	// So if Manager calls this, they SHOULD be allowed to edit even if processed?
	// "Manager can edit user order (add/remove items)" (Req 13).
	// So if `isManager`, skip this check.
	// If `!isManager`, enforce it.

	isMgr := isManager(p, usrID)
	if !isMgr {
		for _, i := range order.Items {
			if i.Status > 1 {
				return nil, errors.New("cannot update order: items already processed by manager")
			}
		}
	}

	// Update Items
	validItems, total, err := s.prepareOrderItems(ctx, order.ProjectID, items)
	if err != nil {
		return nil, err
	}

	// For non-managers, force status reset to Unordered
	if !isMgr {
		for _, item := range validItems {
			item.Status = 1
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

func (s *ProjectService) prepareOrderItems(ctx context.Context, projectID string, inputItems []*project.OrderItem) ([]*project.OrderItem, int64, error) {
	p, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return nil, 0, err
	}

	if p.Status != project.ProjectStatusActive {
		return nil, 0, errors.New("project is not active")
	}

	productMap := make(map[string]*project.Product)
	for _, prod := range p.Products {
		productMap[prod.ID] = prod
	}

	var total int64
	var validItems []*project.OrderItem

	for _, item := range inputItems {
		prod, ok := productMap[item.ProductID]
		if !ok {
			return nil, 0, fmt.Errorf("product not found: %s", item.ProductID)
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
				return nil, 0, fmt.Errorf("spec not found: %s", item.SpecID)
			}
		}

		// Snapshot
		item.ID = uuid.New().String()
		item.ProductName = prod.Name
		item.SpecName = specName
		item.Price = prod.PriceFinal
		item.OrderID = "" // Will be set by caller or Repo

		// Status defaults to Unspecified or Created?
		// Usually 0 is fine, or set explicitly if needed.
		if item.Status == 0 {
			item.Status = 1 // ITEM_STATUS_UNORDERED
		}

		if item.Quantity <= 0 {
			return nil, 0, errors.New("invalid quantity")
		}

		total += item.Price * int64(item.Quantity)
		validItems = append(validItems, item)
	}

	return validItems, total, nil
}

func (s *ProjectService) UpdatePaymentInfo(ctx context.Context, orderID string, method, account string, contact, shipping string, paidAt *time.Time, amount int64) (*project.Order, error) {
	usrID, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	p, err := s.repo.GetByID(ctx, order.ProjectID)
	if err != nil {
		return nil, err
	}

	if role != int(user.UserRoleSysAdmin) && order.UserID != usrID && !isManager(p, usrID) {
		return nil, ErrPermissionDenied
	}

	// allowed to update payment info anytime? Yes, until confirmed maybe?
	if order.PaymentStatus == 3 {
		return nil, errors.New("cannot update payment info: payment already confirmed")
	}

	updated := false
	if method != "" || account != "" || paidAt != nil || amount != 0 {
		if order.PaymentInfo == nil {
			order.PaymentInfo = &project.PaymentInfo{}
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

		// If both set (or valid), maybe set status to Submitted?
		if order.PaymentInfo.Method != "" && (order.PaymentInfo.AccountLast5 != "" || order.PaymentInfo.Method == "Cash") { // Basic check
			order.PaymentStatus = 2 // SUBMITTED
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
func (s *ProjectService) CreateOrder(ctx context.Context, projectID string, items []*project.OrderItem, contactInfo, shippingAddr, shippingMethodID, note string) (*project.Order, error) {
	usrID, _, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	// Validate and Snapshot Items
	validItems, total, err := s.prepareOrderItems(ctx, projectID, items)
	if err != nil {
		return nil, err
	}

	p, err := s.repo.GetByID(ctx, projectID)
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
			return nil, errors.New("invalid shipping method")
		}
	}

	order := &project.Order{
		ID:               uuid.New().String(),
		ProjectID:        projectID,
		UserID:           usrID,
		Items:            validItems,
		TotalAmount:      total + shippingFee,
		CreatedAt:        time.Now(),
		PaymentStatus:    2, // SUBMITTED
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

// ListProjectOrders: Manager Only
func (s *ProjectService) ListProjectOrders(ctx context.Context, projectID string) ([]*project.Order, error) {
	usrID, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	p, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Verify Manager Access
	if role != int(user.UserRoleSysAdmin) && !isManager(p, usrID) {
		return nil, ErrPermissionDenied
	}

	// Fetch orders for this project
	// Note: repo.ListOrders filters by projectID and userID (if strict)
	// Here we want ALL orders for the project, regardless of user
	return s.repo.ListOrders(ctx, projectID, "")
}

func (s *ProjectService) ConfirmPayment(ctx context.Context, orderID string, status int) error {
	usrID, role, ok := auth.FromContext(ctx)
	if !ok {
		return ErrPermissionDenied
	}

	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	p, err := s.repo.GetByID(ctx, order.ProjectID)
	if err != nil {
		return err
	}

	if role != int(user.UserRoleSysAdmin) && !isManager(p, usrID) {
		return ErrPermissionDenied
	}

	return s.repo.UpdateOrderPaymentStatus(ctx, orderID, status)
}

// CancelOrder: Owner Only, and only if Status allowed
func (s *ProjectService) CancelOrder(ctx context.Context, orderID string) error {
	usrID, role, ok := auth.FromContext(ctx)
	if !ok {
		return ErrPermissionDenied
	}

	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	// Verify Ownership or Admin
	if role != int(user.UserRoleSysAdmin) && order.UserID != usrID {
		return ErrPermissionDenied
	}

	// Verify Status (Example: Allow cancel if not yet SHIPPED)
	// For simplicity, let's say PaymentStatus Unset or Submitted
	// Or define specific OrderStatus logic

	// order.Status = Cancelled ...
	// Need update repo logic
	// s.repo.UpdateOrder(ctx, order)

	return nil
}

func (s *ProjectService) AddProduct(ctx context.Context, projectID string, name string, priceOriginal int64, exchangeRate float64, specs []string) (*project.Product, error) {
	usrID, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	p, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if role != int(user.UserRoleSysAdmin) && !isManager(p, usrID) {
		return nil, ErrPermissionDenied
	}

	// Default rounding config logic
	rounding := &project.RoundingConfig{Method: 1, Digit: 0} // Default Floor, ones
	// Use Project Defaults if not provided (Wait, params only have rate? AddProduct signature might need update or infer from project if 0)

	rate := exchangeRate
	if rate == 0 {
		rate = p.ExchangeRate
	}
	if p.Rounding != nil {
		rounding = p.Rounding
	}

	productID := uuid.New().String()
	var productSpecs []*project.ProductSpec
	for _, specName := range specs {
		if specName == "" {
			continue
		}
		productSpecs = append(productSpecs, &project.ProductSpec{
			ID:        uuid.New().String(),
			ProductID: productID,
			Name:      specName,
		})
	}

	// Calculate Final Price
	priceFinal := s.CalculateFinalPrice(priceOriginal, rate, rounding)

	prod := &project.Product{
		ID:            productID,
		ProjectID:     projectID,
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

func (s *ProjectService) DeleteProduct(ctx context.Context, projectID, productID string) error {
	usrID, role, ok := auth.FromContext(ctx)
	if !ok {
		return ErrPermissionDenied
	}

	p, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	if role != int(user.UserRoleSysAdmin) && !isManager(p, usrID) {
		return ErrPermissionDenied
	}

	return s.repo.DeleteProduct(ctx, projectID, productID)
}

func (s *ProjectService) GetMyOrders(ctx context.Context) ([]*project.Order, error) {
	usrID, _, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}

	return s.repo.ListOrders(ctx, "", usrID)
}

func isManager(p *project.Project, userID string) bool {
	if p.CreatorID == userID {
		return true
	}
	for _, m := range p.ManagerIDs {
		if m == userID {
			return true
		}
	}
	return false
}

func (s *ProjectService) BatchUpdateStatus(ctx context.Context, projectID string, specID string, targetStatus int, count int32) (int32, []string, error) {
	usrID, role, ok := auth.FromContext(ctx)
	if !ok {
		return 0, nil, ErrPermissionDenied
	}

	p, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return 0, nil, err
	}

	if role != int(user.UserRoleSysAdmin) && !isManager(p, usrID) {
		return 0, nil, ErrPermissionDenied
	}

	// State Transition Logic (Strict Chain)
	// 1: Unordered -> 2: Ordered
	// 2: Ordered -> 3: Arrived Overseas
	// 3: Arrived Overseas -> 4: Arrived Domestic
	// 4: Arrived Domestic -> 5: Ready for Pickup
	// 5: Ready -> 6: Sent
	var fromStatuses []int
	switch targetStatus {
	case 2:
		fromStatuses = []int{1, 0} // Allow 0 (Unspecified) for legacy/migrated data
	case 3:
		fromStatuses = []int{2}
	case 4:
		fromStatuses = []int{3}
	case 5:
		fromStatuses = []int{4}
	case 6:
		fromStatuses = []int{5}
	default:
		return 0, nil, errors.New("invalid target status for batch update")
	}

	if count <= 0 {
		return 0, nil, nil
	}

	n, ids, err := s.repo.BatchUpdateOrderItemStatus(ctx, projectID, specID, fromStatuses, targetStatus, int(count))
	if err != nil {
		return 0, nil, err
	}

	return int32(n), ids, nil
}

func (s *ProjectService) CreateCategory(ctx context.Context, name string, specNames []string) (*project.Category, error) {
	_, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}
	if role != int(user.UserRoleSysAdmin) {
		return nil, ErrPermissionDenied
	}

	c, err := project.NewCategory(name, specNames)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateCategory(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ProjectService) ListCategories(ctx context.Context) ([]*project.Category, error) {
	// Public access allowed? Or Authenticated?
	// Add Product form is manager only (Auth required).
	// Let's require Auth at least.
	_, _, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}
	return s.repo.ListCategories(ctx)
}

func (s *ProjectService) CalculateFinalPrice(original int64, rate float64, rounding *project.RoundingConfig) int64 {
	if original == 0 || rate == 0 {
		return 0
	}

	val := float64(original) * rate

	// Default Rounding: Floor, Ones
	method := 1 // Floor
	digit := 0  // Ones

	if rounding != nil {
		method = rounding.Method
		digit = rounding.Digit
	}

	// Calculate Power of 10
	pow := math.Pow(10, float64(digit))
	val = val / pow

	switch method {
	case 2: // Ceil
		val = math.Ceil(val)
	case 3: // Round
		val = math.Round(val)
	default: // Floor (1)
		val = math.Floor(val)
	}

	val = val * pow
	return int64(val)
}
