package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/buygo/buygo-api/internal/domain/event"
	"github.com/buygo/buygo-api/internal/domain/user"
)

type Event struct {
	ID                   string `gorm:"primaryKey"`
	Title                string
	Description          string
	CoverImage           string `gorm:"column:cover_image_url"`
	Status               int
	StartTime            time.Time
	EndTime              time.Time
	RegistrationDeadline time.Time
	Location             string
	CreatorID            string
	Creator              *User           `gorm:"foreignKey:CreatorID"`
	Managers             []*User         `gorm:"many2many:event_managers;"`
	PaymentMethods       pq.StringArray  `gorm:"type:text[]"`
	Items                []*EventItem    `gorm:"foreignKey:EventID"`
	Discounts            []*DiscountRule `gorm:"foreignKey:EventID"`
	AllowException       bool            `gorm:"column:allow_modification"`
}

type DiscountRule struct {
	ID               string `gorm:"primaryKey"`
	EventID          string
	MinQuantity      int32
	MinDistinctItems int32
	DiscountAmount   int64
}

type EventItem struct {
	ID              string `gorm:"primaryKey"`
	EventID         string
	Name            string
	Price           int64
	MinParticipants int32
	MaxParticipants int32
	StartTime       *time.Time
	EndTime         *time.Time
	AllowMultiple   bool `gorm:"column:allow_multiple"`
}

type Registration struct {
	ID              string `gorm:"primaryKey"`
	EventID         string
	UserID          string
	Status          int
	PaymentStatus   int
	ContactInfo     string
	Notes           string
	TotalAmount     int64
	DiscountApplied int64
	SelectedItems   []*RegistrationItem `gorm:"foreignKey:RegistrationID"`
	User            *User               `gorm:"foreignKey:UserID"`
}

type RegistrationItem struct {
	ID             string `gorm:"primaryKey"` // Need ID for DB even if not in Domain explicitly? Domain uses value object style?
	RegistrationID string
	EventItemID    string
	Quantity       int
}

// Mappers

func (e *Event) ToDomain() *event.Event {
	var managers []*user.User
	for _, m := range e.Managers {
		managers = append(managers, m.ToDomain())
	}
	var managerIDs []string
	for _, m := range e.Managers {
		managerIDs = append(managerIDs, m.ID)
	}

	var items []*event.EventItem
	for _, i := range e.Items {
		items = append(items, i.ToDomain())
	}

	var discounts []*event.DiscountRule
	for _, d := range e.Discounts {
		discounts = append(discounts, d.ToDomain())
	}

	return &event.Event{
		ID:                   e.ID,
		Title:                e.Title,
		Description:          e.Description,
		CoverImage:           e.CoverImage,
		Status:               event.EventStatus(e.Status),
		StartTime:            e.StartTime,
		EndTime:              e.EndTime,
		RegistrationDeadline: e.RegistrationDeadline,
		Location:             e.Location,
		CreatorID:            e.CreatorID,
		Creator:              e.Creator.ToDomainValid(),
		ManagerIDs:           managerIDs,
		Managers:             managers,
		PaymentMethods:       e.PaymentMethods,
		Items:                items,
		Discounts:            discounts,
		AllowException:       e.AllowException,
	}
}

func FromDomainEvent(e *event.Event) *Event {
	var managers []*User
	for _, id := range e.ManagerIDs {
		managers = append(managers, &User{ID: id})
	}
	var items []*EventItem
	for _, i := range e.Items {
		items = append(items, FromDomainEventItem(i))
	}
	var discounts []*DiscountRule
	for _, d := range e.Discounts {
		discounts = append(discounts, FromDomainDiscountRule(d))
	}

	return &Event{
		ID:                   e.ID,
		Title:                e.Title,
		Description:          e.Description,
		CoverImage:           e.CoverImage,
		Status:               int(e.Status),
		StartTime:            e.StartTime,
		EndTime:              e.EndTime,
		RegistrationDeadline: e.RegistrationDeadline,
		Location:             e.Location,
		CreatorID:            e.CreatorID,
		Managers:             managers,
		PaymentMethods:       e.PaymentMethods,
		Items:                items,
		Discounts:            discounts,
		AllowException:       e.AllowException,
	}
}

// Sub Mappers
func (d *DiscountRule) ToDomain() *event.DiscountRule {
	return &event.DiscountRule{
		MinQuantity:      d.MinQuantity,
		MinDistinctItems: d.MinDistinctItems,
		DiscountAmount:   d.DiscountAmount,
	}
}

func FromDomainDiscountRule(d *event.DiscountRule) *DiscountRule {
	return &DiscountRule{
		ID:               uuid.New().String(), // Generate ID since domain doesn't have it
		MinQuantity:      d.MinQuantity,
		MinDistinctItems: d.MinDistinctItems,
		DiscountAmount:   d.DiscountAmount,
	}
}

func (i *EventItem) ToDomain() *event.EventItem {
	return &event.EventItem{
		ID:              i.ID,
		EventID:         i.EventID,
		Name:            i.Name,
		Price:           i.Price,
		MinParticipants: i.MinParticipants,
		MaxParticipants: i.MaxParticipants,
		StartTime:       i.StartTime,
		EndTime:         i.EndTime,
		AllowMultiple:   i.AllowMultiple,
	}
}

func FromDomainEventItem(i *event.EventItem) *EventItem {
	return &EventItem{
		ID:              i.ID,
		EventID:         i.EventID,
		Name:            i.Name,
		Price:           i.Price,
		MinParticipants: i.MinParticipants,
		MaxParticipants: i.MaxParticipants,
		StartTime:       i.StartTime,
		EndTime:         i.EndTime,
		AllowMultiple:   i.AllowMultiple,
	}
}

// Registration Mappers
func (r *Registration) ToDomain() *event.Registration {
	var items []*event.RegistrationItem
	for _, i := range r.SelectedItems {
		items = append(items, &event.RegistrationItem{
			EventItemID: i.EventItemID, Quantity: i.Quantity,
		})
	}
	var u *user.User
	if r.User != nil {
		u = r.User.ToDomain()
	}

	return &event.Registration{
		ID:              r.ID,
		EventID:         r.EventID,
		UserID:          r.UserID,
		Status:          event.RegistrationStatus(r.Status),
		PaymentStatus:   event.PaymentStatus(r.PaymentStatus),
		ContactInfo:     r.ContactInfo,
		Notes:           r.Notes,
		TotalAmount:     r.TotalAmount,
		DiscountApplied: r.DiscountApplied,
		SelectedItems:   items,
		User:            u,
	}
}

func FromDomainRegistration(r *event.Registration) *Registration {
	var items []*RegistrationItem
	for _, i := range r.SelectedItems {
		items = append(items, &RegistrationItem{
			ID:             uuid.New().String(),
			RegistrationID: r.ID,
			EventItemID:    i.EventItemID,
			Quantity:       i.Quantity,
		})
	}
	// Note: We don't usually set User from Domain to DB during create,
	// as UserID is enough. But if needed for some reason, could map it.
	// For now, assume User association is read-only or managed via UserID.
	return &Registration{
		ID:              r.ID,
		EventID:         r.EventID,
		UserID:          r.UserID,
		Status:          int(r.Status),
		PaymentStatus:   int(r.PaymentStatus),
		ContactInfo:     r.ContactInfo,
		Notes:           r.Notes,
		TotalAmount:     r.TotalAmount,
		DiscountApplied: r.DiscountApplied,
		SelectedItems:   items,
	}
}
