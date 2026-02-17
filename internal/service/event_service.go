package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

type EventService struct {
	repo event.Repository
}

func NewEventService(repo event.Repository) *EventService {
	return &EventService{
		repo: repo,
	}
}

// CreateEvent: Creator/Admin Only
func (s *EventService) CreateEvent(ctx context.Context, title, description, location, coverImage string, start, end time.Time, registrationDeadline *time.Time, paymentMethods []string, allowModification bool, managerIDs []string, items []*event.EventItem, discounts []*event.DiscountRule) (*event.Event, error) {
	usrID, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	if _, _, err := requireRole(ctx, user.UserRoleCreator); err != nil {
		return nil, err
	}

	eventID := uuid.New().String()
	for _, item := range items {
		if item.ID == "" {
			item.ID = uuid.New().String()
		}
		item.EventID = eventID
	}

	mgrs := append([]string{usrID}, managerIDs...)
	// Deduplicate manager IDs
	seen := make(map[string]bool)
	uniqueMgrs := make([]string, 0, len(mgrs))
	for _, m := range mgrs {
		if !seen[m] {
			seen[m] = true
			uniqueMgrs = append(uniqueMgrs, m)
		}
	}

	e := &event.Event{
		ID:             eventID,
		Title:          title,
		Description:    description,
		Location:       location,
		CoverImage:     coverImage,
		StartTime:      start,
		EndTime:        end,
		CreatorID:      usrID,
		Status:         event.EventStatusDraft,
		ManagerIDs:     uniqueMgrs,
		PaymentMethods: paymentMethods,
		AllowException: allowModification,
		Discounts:      discounts,
		Items:          items,
		CreatedAt:      time.Now(),
	}
	if registrationDeadline != nil {
		e.RegistrationDeadline = *registrationDeadline
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

// ListEvents: Public Only
func (s *EventService) ListEvents(ctx context.Context, limit, offset int) ([]*event.Event, error) {
	// Always Public View
	return s.repo.List(ctx, limit, offset, "", false, false)
}

// ListManagerEvents: Authenticated Manager/Admin View
func (s *EventService) ListManagerEvents(ctx context.Context, limit, offset int) ([]*event.Event, error) {
	userID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}
	isSysAdmin := role == int(user.UserRoleSysAdmin)
	return s.repo.List(ctx, limit, offset, userID, isSysAdmin, true)
}

// GetEvent: Public
func (s *EventService) GetEvent(ctx context.Context, id string) (*event.Event, error) {
	return s.repo.GetByID(ctx, id)
}

// RegisterEvent: User Only
func (s *EventService) RegisterEvent(ctx context.Context, eventID string, items []*event.RegistrationItem, contactInfo, notes string) (*event.Registration, error) {
	usrID, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	// Check Event Status
	e, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if e.Status != event.EventStatusActive {
		return nil, ErrNotActive
	}

	// Double check deadline (optional but good practice)
	if !e.RegistrationDeadline.IsZero() && time.Now().After(e.RegistrationDeadline) {
		return nil, ErrDeadlinePassed
	}

	// Check if already registered
	existingRegs, err := s.repo.ListRegistrations(ctx, eventID, usrID)
	if err != nil {
		return nil, err
	}
	// If any active registration exists, error out or return it.
	for _, er := range existingRegs {
		if er.Status != event.RegistrationStatusCancelled {
			return nil, ErrAlreadyRegistered
		}
	}

	// Check item limits
	for _, regItem := range items {
		// Find event item
		var evtItem *event.EventItem
		for _, ei := range e.Items {
			if ei.ID == regItem.EventItemID {
				evtItem = ei
				break
			}
		}
		if evtItem == nil {
			return nil, ErrInvalidEventItemID
		}
		if regItem.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}
		if !evtItem.AllowMultiple && regItem.Quantity > 1 {
			return nil, fmt.Errorf("%s: %w", evtItem.Name, ErrQuantityLimitExceeded)
		}
	}

	reg := &event.Registration{
		ID:            uuid.New().String(),
		EventID:       eventID,
		UserID:        usrID,
		Status:        event.RegistrationStatusPending,
		ContactInfo:   contactInfo,
		Notes:         notes,
		SelectedItems: items,
		PaymentStatus: event.PaymentStatusUnpaid,
	}

	// Calculate Totals & Discounts
	reg.TotalAmount, reg.DiscountApplied = s.calculateTotal(e, items)

	if err := s.repo.Register(ctx, reg); err != nil {
		return nil, err
	}
	return reg, nil
}

// UpdateRegistration: User Only (with checks)
func (s *EventService) UpdateRegistration(ctx context.Context, regID string, items []*event.RegistrationItem, contactInfo, notes string) (*event.Registration, error) {
	usrID, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	reg, err := s.repo.GetRegistration(ctx, regID)
	if err != nil {
		return nil, err
	}

	if reg.UserID != usrID {
		return nil, ErrPermissionDenied
	}

	e, err := s.repo.GetByID(ctx, reg.EventID)
	if err != nil {
		return nil, err
	}

	// Check modification rules
	// 1. If AllowException is true, allow modification regardless of deadline?
	//    Usually exceptions allow overriding deadline, but let's say "AllowModification" means "User can modify".
	//    If False, user cannot modify at all once submitted? Or just not after deadline?
	//    Let's interpret: AllowException=True => User can ALWAYS modify. AllowException=False => User can modify ONLY if before deadline.
	//    Wait, previous context implies "AllowException" means "Allow modification even if... something".
	//    Re-reading requirement: "Event-level control over registration modification (AllowException)."
	//    Let's go with:
	//    - If Deadline Passed AND AllowException False -> Error
	//    - If Deadline Passed AND AllowException True -> OK (Exception allowed)
	//    - If Deadline Not Passed -> OK (Normal modification)

	deadlinePassed := !e.RegistrationDeadline.IsZero() && time.Now().After(e.RegistrationDeadline)

	if deadlinePassed && !e.AllowException {
		return nil, ErrModificationDenied
	}

	// If AllowException is strictly about "Allowing ANY modification", then:
	// if !e.AllowException { return Error } -> This would mean NO edits ever. Too strict.
	// Let's assume standard flow: Edits allowed before deadline. Exception allows edits AFTER deadline.

	// Check item limits for new items
	for _, regItem := range items {
		// Find event item
		var evtItem *event.EventItem
		for _, ei := range e.Items {
			if ei.ID == regItem.EventItemID {
				evtItem = ei
				break
			}
		}
		if evtItem == nil {
			return nil, ErrInvalidEventItemID
		}
		if !evtItem.AllowMultiple && regItem.Quantity > 1 {
			return nil, fmt.Errorf("%s: %w", evtItem.Name, ErrQuantityLimitExceeded)
		}
	}

	// Smart Status Reset Logic
	itemsChanged := false
	if len(reg.SelectedItems) != len(items) {
		itemsChanged = true
	} else {
		// Map old items for easy lookup
		oldMap := make(map[string]int)
		for _, i := range reg.SelectedItems {
			oldMap[i.EventItemID] = i.Quantity
		}
		for _, i := range items {
			if oldQty, exists := oldMap[i.EventItemID]; !exists || oldQty != i.Quantity {
				itemsChanged = true
				break
			}
		}
	}

	if itemsChanged && reg.Status == event.RegistrationStatusConfirmed {
		reg.Status = event.RegistrationStatusPending
	}

	reg.SelectedItems = items
	reg.ContactInfo = contactInfo
	reg.Notes = notes

	// Recalculate Totals
	reg.TotalAmount, reg.DiscountApplied = s.calculateTotal(e, items)

	if err := s.repo.UpdateRegistration(ctx, reg); err != nil {
		return nil, err
	}
	return reg, nil
}

// UpdateRegistrationStatus: Manager/Admin Only
func (s *EventService) UpdateRegistrationStatus(ctx context.Context, regID string, status event.RegistrationStatus, paymentStatus event.PaymentStatus) (*event.Registration, error) {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	reg, err := s.repo.GetRegistration(ctx, regID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	// Must be Admin OR Event Manager
	if !isSysAdmin(role) {
		e, err := s.repo.GetByID(ctx, reg.EventID)
		if err != nil {
			return nil, err
		}
		if !canManage(role, usrID, e) {
			return nil, ErrPermissionDenied
		}
	}

	if status != event.RegistrationStatusUnspecified {
		reg.Status = status
	}
	if paymentStatus != event.PaymentStatusUnspecified {
		reg.PaymentStatus = paymentStatus
	}

	if err := s.repo.UpdateRegistration(ctx, reg); err != nil {
		return nil, err
	}
	return reg, nil
}

// CancelRegistration: Owner Only
func (s *EventService) CancelRegistration(ctx context.Context, regID string) error {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return err
	}

	reg, err := s.repo.GetRegistration(ctx, regID)
	if err != nil {
		return err
	}

	if role != int(user.UserRoleSysAdmin) && reg.UserID != usrID {
		return ErrPermissionDenied
	}

	reg.Status = event.RegistrationStatusCancelled
	return s.repo.UpdateRegistration(ctx, reg)
}

// GetMyRegistrations: User Only
func (s *EventService) GetMyRegistrations(ctx context.Context) ([]*event.Registration, error) {
	usrID, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListRegistrations(ctx, "", usrID)
}

// ListEventRegistrations: Manager/Admin Only
func (s *EventService) ListEventRegistrations(ctx context.Context, eventID string) ([]*event.Registration, error) {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	e, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !canManage(role, usrID, e) {
		return nil, ErrPermissionDenied
	}

	return s.repo.ListRegistrations(ctx, eventID, "")
}

// UpdateEvent: Manager/Admin Only
func (s *EventService) UpdateEvent(ctx context.Context, id string, title, desc, location, cover string, start, end time.Time, allowMod bool, items []*event.EventItem, managerIDs []string, discounts []*event.DiscountRule) (*event.Event, error) {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Permission Check
	if !canManage(role, usrID, e) {
		return nil, ErrPermissionDenied
	}

	// Update Fields
	e.Title = title
	e.Description = desc
	e.Location = location
	e.CoverImage = cover
	e.StartTime = start
	e.EndTime = end
	e.AllowException = allowMod
	// Update Discounts (Full Replace)
	e.Discounts = discounts

	// Update Managers (Security: Only Creator or SysAdmin)
	if managerIDs != nil {
		if role == int(user.UserRoleSysAdmin) || e.CreatorID == usrID {
			e.ManagerIDs = managerIDs
		}
	}

	// Update Items
	// Ensure IDs are preserved if passed, or new IDs generated if empty
	for _, item := range items {
		if item.ID == "" {
			item.ID = uuid.New().String()
		}
		item.EventID = e.ID
		// Basic validation could happen here
	}
	e.Items = items

	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}

	return e, nil
}

// UpdateEventStatus: Manager/Admin Only
func (s *EventService) UpdateEventStatus(ctx context.Context, id string, status event.EventStatus) (*event.Event, error) {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Permission Check
	if !canManage(role, usrID, e) {
		return nil, ErrPermissionDenied
	}

	e.Status = status
	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}

	return e, nil
}

// calculateTotal computes the subtotal and best applicable discount for the given items.
func (s *EventService) calculateTotal(e *event.Event, items []*event.RegistrationItem) (int64, int64) {
	var subtotal int64
	var totalQty int64
	distinctItems := make(map[string]bool)

	// Map items for price lookup
	priceMap := make(map[string]int64)
	for _, ei := range e.Items {
		priceMap[ei.ID] = ei.Price
	}

	for _, item := range items {
		price := priceMap[item.EventItemID]
		subtotal += price * int64(item.Quantity)
		totalQty += int64(item.Quantity)
		if item.Quantity > 0 {
			distinctItems[item.EventItemID] = true
		}
	}

	distinctCount := int64(len(distinctItems))

	// Apply Best Discount Rule
	// Criteria:
	// 1. Total Qty >= Rule.MinQuantity
	// 2. Distinct Count >= Rule.MinDistinctItems
	var maxDiscount int64
	for _, rule := range e.Discounts {
		if totalQty >= int64(rule.MinQuantity) && distinctCount >= int64(rule.MinDistinctItems) {
			if rule.DiscountAmount > maxDiscount {
				maxDiscount = rule.DiscountAmount
			}
		}
	}

	// Ensure discount doesn't exceed total
	if maxDiscount > subtotal {
		maxDiscount = subtotal
	}

	return subtotal - maxDiscount, maxDiscount
}
