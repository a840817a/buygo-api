package event

import (
	"context"
	"time"

	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

type EventStatus int

const (
	EventStatusUnspecified EventStatus = 0
	EventStatusDraft       EventStatus = 1
	EventStatusActive      EventStatus = 2
	EventStatusEnded       EventStatus = 3
	EventStatusArchived    EventStatus = 4
)

type RegistrationStatus int

const (
	RegistrationStatusUnspecified RegistrationStatus = 0
	RegistrationStatusPending     RegistrationStatus = 1
	RegistrationStatusConfirmed   RegistrationStatus = 2
	RegistrationStatusCancelled   RegistrationStatus = 3
)

type PaymentStatus int

const (
	PaymentStatusUnspecified PaymentStatus = 0
	PaymentStatusUnpaid      PaymentStatus = 1
	PaymentStatusSubmitted   PaymentStatus = 2
	PaymentStatusPaid        PaymentStatus = 3
	PaymentStatusRefunded    PaymentStatus = 4
)

type Event struct {
	ID                   string
	Title                string
	Description          string
	CoverImage           string
	Status               EventStatus
	StartTime            time.Time
	EndTime              time.Time
	RegistrationDeadline time.Time
	Location             string
	CreatorID            string
	Creator              *user.User
	ManagerIDs           []string
	Managers             []*user.User
	PaymentMethods       []string
	Items                []*EventItem
	AllowException       bool // Allow modification after registration
	Discounts            []*DiscountRule
	CreatedAt            time.Time
}

type DiscountRule struct {
	MinQuantity      int32
	MinDistinctItems int32
	DiscountAmount   int64
}

type EventItem struct {
	ID              string
	EventID         string
	Name            string
	Price           int64
	MinParticipants int32
	MaxParticipants int32
	StartTime       *time.Time
	EndTime         *time.Time
	AllowMultiple   bool
}

type Registration struct {
	ID              string
	EventID         string
	UserID          string
	Status          RegistrationStatus
	PaymentStatus   PaymentStatus
	ContactInfo     string
	Notes           string
	SelectedItems   []*RegistrationItem
	TotalAmount     int64
	DiscountApplied int64
	User            *user.User
}

type RegistrationItem struct {
	EventItemID string
	Quantity    int
}

// IsManager checks if the given user is the creator or a manager of this event.
func (e *Event) IsManager(userID string) bool {
	return user.CheckIsManager(e.CreatorID, e.ManagerIDs, userID)
}

type Repository interface {
	Create(ctx context.Context, e *Event) error
	GetByID(ctx context.Context, id string) (*Event, error)
	List(ctx context.Context, limit int, offset int, userID string, isSysAdmin bool, manageOnly bool) ([]*Event, error)
	Update(ctx context.Context, e *Event) error

	Register(ctx context.Context, r *Registration) error
	GetRegistration(ctx context.Context, id string) (*Registration, error)
	ListRegistrations(ctx context.Context, eventID string, userID string) ([]*Registration, error)
	UpdateRegistration(ctx context.Context, r *Registration) error
}
