package service

import (
	"context"

	"github.com/buygo/buygo-api/internal/domain/groupbuy"
	"github.com/buygo/buygo-api/internal/domain/user"
)

// CreatePriceTemplate: Admin Only
func (s *GroupBuyService) CreatePriceTemplate(ctx context.Context, name, sourceCurrency string, rate float64, rounding *groupbuy.RoundingConfig) (*groupbuy.PriceTemplate, error) {
	if _, _, err := requireRole(ctx, user.UserRoleSysAdmin); err != nil {
		return nil, err
	}

	pt := groupbuy.NewPriceTemplate(name, sourceCurrency, rate, rounding)
	if err := s.repo.CreatePriceTemplate(ctx, pt); err != nil {
		return nil, err
	}

	return pt, nil
}

// ListPriceTemplates: Authenticated (Managers need to see them to select)
func (s *GroupBuyService) ListPriceTemplates(ctx context.Context) ([]*groupbuy.PriceTemplate, error) {
	if _, _, err := checkLogin(ctx); err != nil {
		return nil, err
	}
	return s.repo.ListPriceTemplates(ctx)
}

// GetPriceTemplate: Authenticated
func (s *GroupBuyService) GetPriceTemplate(ctx context.Context, id string) (*groupbuy.PriceTemplate, error) {
	if _, _, err := checkLogin(ctx); err != nil {
		return nil, err
	}
	return s.repo.GetPriceTemplate(ctx, id)
}

// UpdatePriceTemplate: Admin Only
func (s *GroupBuyService) UpdatePriceTemplate(ctx context.Context, id, name, sourceCurrency string, rate float64, rounding *groupbuy.RoundingConfig) (*groupbuy.PriceTemplate, error) {
	if _, _, err := requireRole(ctx, user.UserRoleSysAdmin); err != nil {
		return nil, err
	}

	pt, err := s.repo.GetPriceTemplate(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		pt.Name = name
	}
	if sourceCurrency != "" {
		pt.SourceCurrency = sourceCurrency
	}
	if rate != 0 {
		pt.ExchangeRate = rate
	}
	if rounding != nil {
		pt.Rounding = rounding
	}
	// Update timestamp? Usually repo handles or we do here.
	// pt.UpdatedAt = time.Now()

	if err := s.repo.UpdatePriceTemplate(ctx, pt); err != nil {
		return nil, err
	}

	return pt, nil
}

// DeletePriceTemplate: Admin Only
func (s *GroupBuyService) DeletePriceTemplate(ctx context.Context, id string) error {
	if _, _, err := requireRole(ctx, user.UserRoleSysAdmin); err != nil {
		return err
	}

	return s.repo.DeletePriceTemplate(ctx, id)
}
