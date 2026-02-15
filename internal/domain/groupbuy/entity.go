package groupbuy

import (
	"context"
	"time"

	"github.com/buygo/buygo-api/internal/domain/user"
)

type GroupBuyStatus int

const (
	GroupBuyStatusUnspecified GroupBuyStatus = 0
	GroupBuyStatusDraft       GroupBuyStatus = 1
	GroupBuyStatusActive      GroupBuyStatus = 2
	GroupBuyStatusEnded       GroupBuyStatus = 3
	GroupBuyStatusArchived    GroupBuyStatus = 4
)

type GroupBuy struct {
	ID           string
	Title        string
	Description  string
	CoverImage   string
	Status       GroupBuyStatus
	ExchangeRate float64
	Rounding     *RoundingConfig
	CreatorID    string
	Creator      *user.User
	ManagerIDs   []string
	Managers     []*user.User

	SourceCurrency string // e.g. "JPY", "USD"

	ShippingConfigs []*ShippingConfig
	Products        []*Product
	CreatedAt       time.Time
	Deadline        *time.Time
}

type Product struct {
	ID            string
	GroupBuyID    string
	Name          string
	Description   string
	ImageURL      string
	PriceOriginal int64
	ExchangeRate  float64
	PriceFinal    int64
	MaxQuantity   int32
	Rounding      *RoundingConfig
	Specs         []*ProductSpec
}

type RoundingMethod int

const (
	RoundingMethodUnspecified RoundingMethod = 0
	RoundingMethodFloor       RoundingMethod = 1
	RoundingMethodCeil        RoundingMethod = 2
	RoundingMethodRound       RoundingMethod = 3
)

type RoundingConfig struct {
	Method RoundingMethod
	Digit  int
}

type ProductSpec struct {
	ID        string
	ProductID string
	Name      string
}

type ShippingType int

const (
	ShippingTypeUnspecified ShippingType = 0
	ShippingTypeDelivery    ShippingType = 1
	ShippingTypeStorePickup ShippingType = 2
	ShippingTypeMeetup      ShippingType = 3
)

type ShippingConfig struct {
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Type  ShippingType `json:"type"`
	Price int64        `json:"price"`
}

type OrderItemStatus int

const (
	OrderItemStatusUnspecified     OrderItemStatus = 0
	OrderItemStatusUnordered       OrderItemStatus = 1
	OrderItemStatusOrdered         OrderItemStatus = 2
	OrderItemStatusArrivedOverseas OrderItemStatus = 3
	OrderItemStatusArrivedDomestic OrderItemStatus = 4
	OrderItemStatusReadyForPickup  OrderItemStatus = 5
	OrderItemStatusSent            OrderItemStatus = 6
	OrderItemStatusFailed          OrderItemStatus = 7
)

type PaymentStatus int

const (
	PaymentStatusUnspecified PaymentStatus = 0
	PaymentStatusUnset       PaymentStatus = 1
	PaymentStatusSubmitted   PaymentStatus = 2
	PaymentStatusConfirmed   PaymentStatus = 3
	PaymentStatusRejected    PaymentStatus = 4
)

// Order Entity
type Order struct {
	ID               string
	GroupBuyID       string
	UserID           string
	TotalAmount      int64
	PaymentStatus    PaymentStatus
	ContactInfo      string
	ShippingAddress  string
	PaymentInfo      *PaymentInfo
	Items            []*OrderItem
	ShippingMethodID string
	ShippingFee      int64
	Note             string
	CreatedAt        time.Time
}

type PaymentInfo struct {
	Method       string
	AccountLast5 string
	PaidAt       *time.Time
	Amount       int64
}

type OrderItem struct {
	ID          string
	OrderID     string
	ProductID   string
	SpecID      string
	Quantity    int
	Status      OrderItemStatus
	ProductName string
	SpecName    string
	Price       int64
}

// IsManager checks if the given user is the creator or a manager of this group buy.
func (gb *GroupBuy) IsManager(userID string) bool {
	if gb.CreatorID == userID {
		return true
	}
	for _, m := range gb.ManagerIDs {
		if m == userID {
			return true
		}
	}
	return false
}

// Repository Port
type Repository interface {
	Create(ctx context.Context, gb *GroupBuy) error
	GetByID(ctx context.Context, id string) (*GroupBuy, error)
	List(ctx context.Context, limit int, offset int, userID string, isSysAdmin bool, manageOnly bool) ([]*GroupBuy, error)
	Update(ctx context.Context, gb *GroupBuy) error

	AddProduct(ctx context.Context, product *Product) error
	DeleteProduct(ctx context.Context, groupBuyID, productID string) error

	CreateOrder(ctx context.Context, order *Order) error
	GetOrder(ctx context.Context, id string) (*Order, error)
	ListOrders(ctx context.Context, groupBuyID string, userID string) ([]*Order, error)
	UpdateOrder(ctx context.Context, order *Order) error
	UpdateOrderPaymentStatus(ctx context.Context, orderID string, status PaymentStatus) error
	BatchUpdateOrderItemStatus(ctx context.Context, groupBuyID string, specID string, fromStatuses []int, toStatus int, limit int) (int64, []string, error)

	// Category
	CreateCategory(ctx context.Context, c *Category) error
	ListCategories(ctx context.Context) ([]*Category, error)

	// PriceTemplate
	CreatePriceTemplate(ctx context.Context, pt *PriceTemplate) error
	ListPriceTemplates(ctx context.Context) ([]*PriceTemplate, error)
	GetPriceTemplate(ctx context.Context, id string) (*PriceTemplate, error)
	UpdatePriceTemplate(ctx context.Context, pt *PriceTemplate) error
	DeletePriceTemplate(ctx context.Context, id string) error
}
