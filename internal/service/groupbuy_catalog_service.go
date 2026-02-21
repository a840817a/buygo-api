package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

// AddProduct adds a new product to an existing group buy (manager only).
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

	rounding := &groupbuy.RoundingConfig{Method: groupbuy.RoundingMethodFloor, Digit: 0}

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

// DeleteProduct removes a product from a group buy (manager only).
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

// CreateCategory creates a new product category (sys admin only).
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

// ListCategories returns all product categories (authenticated users).
func (s *GroupBuyService) ListCategories(ctx context.Context) ([]*groupbuy.Category, error) {
	_, _, err := checkLogin(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.ListCategories(ctx)
}
