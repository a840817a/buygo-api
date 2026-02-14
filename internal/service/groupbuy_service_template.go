package service

import (
	"context"

	"github.com/buygo/buygo-api/internal/domain/auth"
	"github.com/buygo/buygo-api/internal/domain/groupbuy"
	"github.com/buygo/buygo-api/internal/domain/user"
)

// CreatePriceTemplate: Admin Only
func (s *GroupBuyService) CreatePriceTemplate(ctx context.Context, name, sourceCurrency string, rate float64, rounding *project.RoundingConfig) (*project.PriceTemplate, error) {
	_, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}
	if role != int(user.UserRoleSysAdmin) {
		return nil, ErrPermissionDenied
	}

	pt := project.NewPriceTemplate(name, sourceCurrency, rate, rounding)
	if err := s.repo.CreatePriceTemplate(ctx, pt); err != nil {
		return nil, err
	}

	return pt, nil
}

// ListPriceTemplates: Authenticated (Managers need to see them to select)
func (s *GroupBuyService) ListPriceTemplates(ctx context.Context) ([]*project.PriceTemplate, error) {
	_, _, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}
	return s.repo.ListPriceTemplates(ctx)
}

// GetPriceTemplate: Authenticated
func (s *GroupBuyService) GetPriceTemplate(ctx context.Context, id string) (*project.PriceTemplate, error) {
	_, _, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}
	return s.repo.GetPriceTemplate(ctx, id)
}

// UpdatePriceTemplate: Admin Only
func (s *GroupBuyService) UpdatePriceTemplate(ctx context.Context, id, name, sourceCurrency string, rate float64, rounding *project.RoundingConfig) (*project.PriceTemplate, error) {
	_, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrPermissionDenied
	}
	if role != int(user.UserRoleSysAdmin) {
		return nil, ErrPermissionDenied
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
	_, role, ok := auth.FromContext(ctx)
	if !ok {
		return ErrPermissionDenied
	}
	if role != int(user.UserRoleSysAdmin) {
		return ErrPermissionDenied
	}

	return s.repo.DeletePriceTemplate(ctx, id)
}
