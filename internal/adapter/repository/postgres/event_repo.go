package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/hatsubosi/buygo-api/internal/adapter/repository/postgres/model"
	"github.com/hatsubosi/buygo-api/internal/domain/event"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

// Event Core
func (r *EventRepository) Create(ctx context.Context, e *event.Event) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := model.FromDomainEvent(e)

		// 1. Create Event (and Items by cascade, excluding Users)
		if err := tx.Omit("Creator", "Managers").Create(m).Error; err != nil {
			return err
		}

		// 2. Associate Managers
		if len(m.Managers) > 0 {
			if err := tx.Model(m).Association("Managers").Replace(m.Managers); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *EventRepository) GetByID(ctx context.Context, id string) (*event.Event, error) {
	var m model.Event
	if err := r.db.WithContext(ctx).
		Preload("Creator").
		Preload("Managers").
		Preload("Items").
		Preload("Discounts").
		First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *EventRepository) List(ctx context.Context, limit, offset int, userID string, isSysAdmin bool, manageOnly bool) ([]*event.Event, error) {
	var models []*model.Event

	query := r.db.WithContext(ctx).
		Limit(limit).Offset(offset).
		Preload("Creator").
		Preload("Managers").
		Preload("Discounts")

	if isSysAdmin {
		// No filter
	} else if userID != "" {
		if manageOnly {
			// Strict Manager View: Only events where user is Creator OR Manager
			query = query.Where(
				r.db.Where("creator_id = ?", userID).
					Or("id IN (?)", r.db.Table("event_managers").Select("event_id").Where("user_id = ?", userID)),
			)
		} else {
			// Public + My Drafts View: (Active/Ended) OR (Creator OR Manager)
			query = query.Where(
				r.db.Where("status IN ?", []int{2, 3}).
					Or("creator_id = ?", userID).
					Or("id IN (?)", r.db.Table("event_managers").Select("event_id").Where("user_id = ?", userID)),
			)
		}
	} else {
		query = query.Where("status IN ?", []int{2, 3})
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	var res []*event.Event
	for _, m := range models {
		res = append(res, m.ToDomain())
	}
	return res, nil
}

func (r *EventRepository) Update(ctx context.Context, e *event.Event) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := model.FromDomainEvent(e)

		// 1. Update Basic Fields
		if err := tx.Model(&model.Event{ID: e.ID}).Updates(map[string]interface{}{
			"title":                 m.Title,
			"description":           m.Description,
			"cover_image_url":       m.CoverImage,
			"start_time":            m.StartTime,
			"end_time":              m.EndTime,
			"registration_deadline": m.RegistrationDeadline,
			"location":              m.Location,
			"allow_modification":    m.AllowException,
			"status":                m.Status,
		}).Error; err != nil {
			return err
		}

		// 2. Replace Items
		if err := tx.Model(m).Association("Items").Replace(m.Items); err != nil {
			return err
		}

		// 3. Replace Discounts
		if err := tx.Model(m).Association("Discounts").Replace(m.Discounts); err != nil {
			return err
		}

		return nil
	})
}

// Registration Methods
func (r *EventRepository) Register(ctx context.Context, reg *event.Registration) error {
	m := model.FromDomainRegistration(reg)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *EventRepository) GetRegistration(ctx context.Context, id string) (*event.Registration, error) {
	var m model.Registration
	if err := r.db.WithContext(ctx).Preload("SelectedItems").Preload("User").First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("registration not found")
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *EventRepository) ListRegistrations(ctx context.Context, eventID string, userID string) ([]*event.Registration, error) {
	var models []*model.Registration
	query := r.db.WithContext(ctx).Preload("SelectedItems").Preload("User")

	if eventID != "" {
		query = query.Where("event_id = ?", eventID)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	var res []*event.Registration
	for _, m := range models {
		res = append(res, m.ToDomain())
	}
	return res, nil
}

func (r *EventRepository) UpdateRegistration(ctx context.Context, reg *event.Registration) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := model.FromDomainRegistration(reg)
		// 1. Update basic fields
		if err := tx.Save(m).Error; err != nil {
			return err
		}
		// 2. Replace Items
		if err := tx.Model(m).Association("SelectedItems").Replace(m.SelectedItems); err != nil {
			return err
		}
		return nil
	})
}
