package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	// AutoMigrate is usually done at startup, but ensuring here or in main
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	m := model.FromDomainUser(u)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	var m model.User
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var m model.User
	if err := r.db.WithContext(ctx).First(&m, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	m := model.FromDomainUser(u)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	var models []*model.User
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, err
	}
	var res []*user.User
	for _, m := range models {
		res = append(res, m.ToDomain())
	}
	return res, nil
}
