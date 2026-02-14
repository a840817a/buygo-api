package postgres

import (
	"context"

	"github.com/buygo/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/buygo/buygo-api/internal/domain/groupbuy"
)

func (r *ProjectRepository) CreatePriceTemplate(ctx context.Context, pt *project.PriceTemplate) error {
	m := model.FromDomainPriceTemplate(pt)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *ProjectRepository) ListPriceTemplates(ctx context.Context) ([]*project.PriceTemplate, error) {
	var models []*model.PriceTemplate
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	var res []*project.PriceTemplate
	for _, m := range models {
		res = append(res, m.ToDomain())
	}
	return res, nil
}

func (r *ProjectRepository) GetPriceTemplate(ctx context.Context, id string) (*project.PriceTemplate, error) {
	var m model.PriceTemplate
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *ProjectRepository) UpdatePriceTemplate(ctx context.Context, pt *project.PriceTemplate) error {
	m := model.FromDomainPriceTemplate(pt)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *ProjectRepository) DeletePriceTemplate(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.PriceTemplate{}, "id = ?", id).Error
}
