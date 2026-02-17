package service

import (
	"context"
	"time"

	"github.com/hatsubosi/buygo-api/internal/domain/event"
	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

// GroupBuyServiceInterface defines the contract for group buy operations.
type GroupBuyServiceInterface interface {
	CreateGroupBuy(ctx context.Context, title, description string) (*groupbuy.GroupBuy, error)
	GetGroupBuy(ctx context.Context, id string) (*groupbuy.GroupBuy, error)
	ListGroupBuys(ctx context.Context, limit, offset int) ([]*groupbuy.GroupBuy, error)
	ListManagerGroupBuys(ctx context.Context, limit, offset int) ([]*groupbuy.GroupBuy, error)
	UpdateGroupBuy(ctx context.Context, id string, title, desc string, status groupbuy.GroupBuyStatus, products []*groupbuy.Product, coverImage string, deadline *time.Time, shippingConfigs []*groupbuy.ShippingConfig, managerIDs []string, exchangeRate float64, rounding *groupbuy.RoundingConfig, sourceCurrency string) (*groupbuy.GroupBuy, error)

	CreateOrder(ctx context.Context, groupBuyID string, items []*groupbuy.OrderItem, contactInfo, shippingAddr, shippingMethodID, note string) (*groupbuy.Order, error)
	GetMyGroupBuyOrder(ctx context.Context, groupBuyID string) (*groupbuy.Order, error)
	UpdateOrder(ctx context.Context, orderID string, items []*groupbuy.OrderItem, note string) (*groupbuy.Order, error)
	UpdatePaymentInfo(ctx context.Context, orderID string, method, account string, contact, shipping string, paidAt *time.Time, amount int64) (*groupbuy.Order, error)
	ListGroupBuyOrders(ctx context.Context, groupBuyID string) ([]*groupbuy.Order, error)
	GetMyOrders(ctx context.Context) ([]*groupbuy.Order, error)
	ConfirmPayment(ctx context.Context, orderID string, status groupbuy.PaymentStatus) error
	CancelOrder(ctx context.Context, orderID string) error
	BatchUpdateStatus(ctx context.Context, groupBuyID string, specID string, targetStatus int, count int32) (int32, []string, error)

	AddProduct(ctx context.Context, groupBuyID string, name string, priceOriginal int64, exchangeRate float64, specs []string) (*groupbuy.Product, error)
	DeleteProduct(ctx context.Context, groupBuyID, productID string) error

	CreateCategory(ctx context.Context, name string, specNames []string) (*groupbuy.Category, error)
	ListCategories(ctx context.Context) ([]*groupbuy.Category, error)

	CreatePriceTemplate(ctx context.Context, name, sourceCurrency string, rate float64, rounding *groupbuy.RoundingConfig) (*groupbuy.PriceTemplate, error)
	ListPriceTemplates(ctx context.Context) ([]*groupbuy.PriceTemplate, error)
	GetPriceTemplate(ctx context.Context, id string) (*groupbuy.PriceTemplate, error)
	UpdatePriceTemplate(ctx context.Context, id, name, sourceCurrency string, rate float64, rounding *groupbuy.RoundingConfig) (*groupbuy.PriceTemplate, error)
	DeletePriceTemplate(ctx context.Context, id string) error

	CalculateFinalPrice(original int64, rate float64, rounding *groupbuy.RoundingConfig) int64
}

// EventServiceInterface defines the contract for event operations.
type EventServiceInterface interface {
	CreateEvent(ctx context.Context, title, description, location, coverImage string, start, end time.Time, registrationDeadline *time.Time, paymentMethods []string, allowModification bool, managerIDs []string, items []*event.EventItem, discounts []*event.DiscountRule) (*event.Event, error)
	ListEvents(ctx context.Context, limit, offset int) ([]*event.Event, error)
	ListManagerEvents(ctx context.Context, limit, offset int) ([]*event.Event, error)
	GetEvent(ctx context.Context, id string) (*event.Event, error)
	UpdateEvent(ctx context.Context, id string, title, desc, location, cover string, start, end time.Time, allowMod bool, items []*event.EventItem, managerIDs []string, discounts []*event.DiscountRule) (*event.Event, error)
	UpdateEventStatus(ctx context.Context, id string, status event.EventStatus) (*event.Event, error)

	RegisterEvent(ctx context.Context, eventID string, items []*event.RegistrationItem, contactInfo, notes string) (*event.Registration, error)
	UpdateRegistration(ctx context.Context, regID string, items []*event.RegistrationItem, contactInfo, notes string) (*event.Registration, error)
	UpdateRegistrationStatus(ctx context.Context, regID string, status event.RegistrationStatus, paymentStatus event.PaymentStatus) (*event.Registration, error)
	CancelRegistration(ctx context.Context, regID string) error
	GetMyRegistrations(ctx context.Context) ([]*event.Registration, error)
	ListEventRegistrations(ctx context.Context, eventID string) ([]*event.Registration, error)
}

// AuthServiceInterface defines the contract for authentication operations.
type AuthServiceInterface interface {
	LoginOrRegister(ctx context.Context, token string) (string, *user.User, error)
	GetMe(ctx context.Context, userID string) (*user.User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*user.User, error)
	UpdateUserRole(ctx context.Context, userID string, role user.UserRole) (*user.User, error)
	ListAssignableManagers(ctx context.Context, query string) ([]*user.User, error)
}

// Compile-time interface compliance checks.
var (
	_ GroupBuyServiceInterface = (*GroupBuyService)(nil)
	_ EventServiceInterface    = (*EventService)(nil)
	_ AuthServiceInterface     = (*AuthService)(nil)
)
