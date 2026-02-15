package postgres

import (
	"context"

	"github.com/buygo/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/buygo/buygo-api/internal/domain/groupbuy"
)

func (r *GroupBuyRepository) CreatePriceTemplate(ctx context.Context, pt *groupbuy.PriceTemplate) error {
	m := model.FromDomainPriceTemplate(pt)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *GroupBuyRepository) ListPriceTemplates(ctx context.Context) ([]*groupbuy.PriceTemplate, error) {
	var models []*model.PriceTemplate
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	var res []*groupbuy.PriceTemplate
	for _, m := range models {
		res = append(res, m.ToDomain())
	}
	return res, nil
}

func (r *GroupBuyRepository) GetPriceTemplate(ctx context.Context, id string) (*groupbuy.PriceTemplate, error) {
	var m model.PriceTemplate
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *GroupBuyRepository) UpdatePriceTemplate(ctx context.Context, pt *groupbuy.PriceTemplate) error {
	m := model.FromDomainPriceTemplate(pt)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *GroupBuyRepository) DeletePriceTemplate(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.PriceTemplate{}, "id = ?", id).Error
}
