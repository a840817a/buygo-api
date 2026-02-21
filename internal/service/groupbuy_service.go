package service

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

// GroupBuyService manages the group buy lifecycle.
type GroupBuyService struct {
	repo groupbuy.Repository
}

// NewGroupBuyService creates a new GroupBuyService backed by the given repository.
func NewGroupBuyService(repo groupbuy.Repository) *GroupBuyService {
	return &GroupBuyService{repo: repo}
}

// CreateGroupBuy creates a new group buy. Requires Creator or Admin role.
func (s *GroupBuyService) CreateGroupBuy(ctx context.Context, title, description string, products []*groupbuy.Product, coverImage string, deadline *time.Time, shippingConfigs []*groupbuy.ShippingConfig, managerIDs []string, exchangeRate float64, rounding *groupbuy.RoundingConfig, sourceCurrency string) (*groupbuy.GroupBuy, error) {
	usrID, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	if _, _, err := requireRole(ctx, user.UserRoleCreator); err != nil {
		return nil, err
	}

	now := time.Now()
	if rounding == nil {
		rounding = &groupbuy.RoundingConfig{Method: groupbuy.RoundingMethodFloor, Digit: 0}
	}
	if exchangeRate == 0 {
		exchangeRate = 0.23
	}
	if sourceCurrency == "" {
		sourceCurrency = "JPY"
	}

	for _, sc := range shippingConfigs {
		if sc.ID == "" {
			sc.ID = uuid.New().String()
		}
	}

	gb := &groupbuy.GroupBuy{
		ID:              uuid.New().String(),
		CreatorID:       usrID,
		Title:           title,
		Description:     description,
		Status:          groupbuy.GroupBuyStatusDraft,
		CoverImage:      coverImage,
		Deadline:        deadline,
		ManagerIDs:      append([]string{usrID}, managerIDs...),
		ShippingConfigs: shippingConfigs,
		ExchangeRate:    exchangeRate,
		SourceCurrency:  sourceCurrency,
		Rounding:        rounding,
		CreatedAt:       now,
	}

	if len(products) > 0 {
		gb.Products = s.prepareProducts(gb.ID, gb.ExchangeRate, gb.Rounding, products)
	}

	if err := s.repo.Create(ctx, gb); err != nil {
		return nil, err
	}
	return gb, nil
}

// GetGroupBuy retrieves a group buy by ID.
func (s *GroupBuyService) GetGroupBuy(ctx context.Context, id string) (*groupbuy.GroupBuy, error) {
	return s.repo.GetByID(ctx, id)
}

// ListGroupBuys returns public group buys with pagination.
func (s *GroupBuyService) ListGroupBuys(ctx context.Context, limit, offset int) ([]*groupbuy.GroupBuy, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.List(ctx, limit, offset, "", false, false)
}

// ListManagerGroupBuys returns group buys visible to the authenticated manager/admin.
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

// UpdateGroupBuy updates an existing group buy. Requires manager, creator, or admin.
func (s *GroupBuyService) UpdateGroupBuy(ctx context.Context, id string, title, desc string, status groupbuy.GroupBuyStatus, products []*groupbuy.Product, coverImage string, deadline *time.Time, shippingConfigs []*groupbuy.ShippingConfig, managerIDs []string, exchangeRate float64, rounding *groupbuy.RoundingConfig, sourceCurrency string) (*groupbuy.GroupBuy, error) {
	usrID, role, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !canManage(role, usrID, p) && p.CreatorID != usrID {
		return nil, ErrPermissionDenied
	}

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

	if managerIDs != nil {
		if role == int(user.UserRoleSysAdmin) || p.CreatorID == usrID {
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

	if len(products) > 0 {
		p.Products = s.prepareProducts(p.ID, p.ExchangeRate, p.Rounding, products)
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

// CalculateFinalPrice converts a source price via exchange rate and rounding.
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

// prepareProducts ensures all products have IDs, correct exchange rates,
// rounding configs, calculated final prices, and spec IDs.
func (s *GroupBuyService) prepareProducts(gbID string, exchangeRate float64, rounding *groupbuy.RoundingConfig, products []*groupbuy.Product) []*groupbuy.Product {
	for _, item := range products {
		if item.ExchangeRate == 0 {
			item.ExchangeRate = exchangeRate
		}
		if item.Rounding == nil {
			item.Rounding = rounding
		}
		if item.ID == "" {
			item.ID = uuid.New().String()
		}
		item.PriceFinal = s.CalculateFinalPrice(item.PriceOriginal, item.ExchangeRate, item.Rounding)
		item.GroupBuyID = gbID

		for _, spec := range item.Specs {
			if spec.ID == "" {
				spec.ID = uuid.New().String()
			}
			spec.ProductID = item.ID
		}
	}
	return products
}
