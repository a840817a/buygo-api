package project

import (
	"context"
	"time"

	"github.com/buygo/buygo-api/internal/domain/user"
)

type ProjectStatus int

const (
	ProjectStatusUnspecified ProjectStatus = 0
	ProjectStatusDraft       ProjectStatus = 1
	ProjectStatusActive      ProjectStatus = 2
	ProjectStatusEnded       ProjectStatus = 3
	ProjectStatusArchived    ProjectStatus = 4
)

type Project struct {
	ID           string
	Title        string
	Description  string
	CoverImage   string
	Status       ProjectStatus
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
	ProjectID     string
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

type RoundingConfig struct {
	Method int // 0: Unspecified, 1: Floor, 2: Ceil, 3: Round
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
} // Project update below

// Order Entity
type Order struct {
	ID               string
	ProjectID        string
	UserID           string
	TotalAmount      int64
	PaymentStatus    int // enum
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
	Status      int // enum
	ProductName string
	SpecName    string
	Price       int64
}

// Repository Port
type Repository interface {
	Create(ctx context.Context, p *Project) error
	GetByID(ctx context.Context, id string) (*Project, error)
	List(ctx context.Context, limit int, offset int, userID string, isSysAdmin bool, manageOnly bool) ([]*Project, error)
	Update(ctx context.Context, p *Project) error

	AddProduct(ctx context.Context, product *Product) error
	DeleteProduct(ctx context.Context, projectID, productID string) error

	CreateOrder(ctx context.Context, order *Order) error
	GetOrder(ctx context.Context, id string) (*Order, error)
	ListOrders(ctx context.Context, projectID string, userID string) ([]*Order, error)
	UpdateOrder(ctx context.Context, order *Order) error
	UpdateOrderPaymentStatus(ctx context.Context, orderID string, status int) error
	BatchUpdateOrderItemStatus(ctx context.Context, projectID string, specID string, fromStatuses []int, toStatus int, limit int) (int64, []string, error)

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
