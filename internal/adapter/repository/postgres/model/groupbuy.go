package model

import (
	"time"

	"github.com/buygo/buygo-api/internal/domain/groupbuy"
	"github.com/buygo/buygo-api/internal/domain/user"
)

type GroupBuy struct {
	ID             string `gorm:"primaryKey"`
	Title          string
	Description    string
	CoverImage     string
	Status         int
	ExchangeRate   float64
	SourceCurrency string
	RoundingMethod int
	RoundingDigit  int
	CreatorID      string
	Creator        *User   `gorm:"foreignKey:CreatorID"`
	Managers       []*User `gorm:"many2many:project_managers;"`

	ShippingConfigs []*groupbuy.ShippingConfig `gorm:"serializer:json"` // Store as JSON
	Products        []*Product                 `gorm:"foreignKey:GroupBuyID"`
	CreatedAt       time.Time
	Deadline        *time.Time
}

type Product struct {
	ID             string `gorm:"primaryKey"`
	GroupBuyID     string
	Name           string
	Description    string
	ImageURL       string
	PriceOriginal  int64
	ExchangeRate   float64
	PriceFinal     int64
	MaxQuantity    int32
	RoundingMethod int
	RoundingDigit  int
	Specs          []*ProductSpec `gorm:"foreignKey:ProductID"`
}

type ProductSpec struct {
	ID        string `gorm:"primaryKey"`
	ProductID string
	Name      string
}

type Order struct {
	ID                  string `gorm:"primaryKey"`
	GroupBuyID          string
	UserID              string
	TotalAmount         int64
	PaymentStatus       int
	ContactInfo         string
	ShippingAddress     string
	PaymentMethod       string
	PaymentAccountLast5 string
	PaidAt              *time.Time
	PaymentAmount       int64
	Items               []*OrderItem `gorm:"foreignKey:OrderID"`
	ShippingMethodID    string
	ShippingFee         int64
	Note                string
	CreatedAt           time.Time
}

type OrderItem struct {
	ID          string `gorm:"primaryKey"`
	OrderID     string
	ProductID   string
	SpecID      string
	Quantity    int
	Status      int
	ProductName string
	SpecName    string
	Price       int64
}

// Mappers

func (gb *GroupBuy) ToDomain() *groupbuy.GroupBuy {
	var managers []*user.User
	for _, m := range gb.Managers {
		managers = append(managers, m.ToDomain())
	}
	var managerIDs []string
	for _, m := range gb.Managers {
		managerIDs = append(managerIDs, m.ID)
	}

	// Products mapper ...
	var products []*groupbuy.Product
	for _, prod := range gb.Products {
		products = append(products, prod.ToDomain())
	}

	return &groupbuy.GroupBuy{
		ID:           gb.ID,
		Title:        gb.Title,
		Description:  gb.Description,
		CoverImage:   gb.CoverImage,
		Status:       groupbuy.GroupBuyStatus(gb.Status),
		ExchangeRate: gb.ExchangeRate,
		Rounding:     &groupbuy.RoundingConfig{Method: groupbuy.RoundingMethod(gb.RoundingMethod), Digit: gb.RoundingDigit},
		CreatorID:    gb.CreatorID,
		Creator:      gb.Creator.ToDomainValid(), // Handle nil
		ManagerIDs:   managerIDs,
		Managers:     managers,

		ShippingConfigs: gb.ShippingConfigs,
		Products:        products,
		CreatedAt:       gb.CreatedAt,
		Deadline:        gb.Deadline,
		SourceCurrency:  gb.SourceCurrency,
	}
}

func FromDomainGroupBuy(gb *groupbuy.GroupBuy) *GroupBuy {
	var managers []*User
	// Note: For create/update, GORM handles association if we provide the struct with ID.
	// We might need to fetch users first or just set IDs references if GORM supports it.
	// Simplest for ManyToMany: Provide full User structs with just ID set is often enough for association update.
	for _, id := range gb.ManagerIDs {
		managers = append(managers, &User{ID: id})
	}

	var products []*Product
	for _, prod := range gb.Products {
		products = append(products, FromDomainProduct(prod))
	}

	var rm, rd int
	if gb.Rounding != nil {
		rm, rd = int(gb.Rounding.Method), gb.Rounding.Digit
	}

	return &GroupBuy{
		ID:             gb.ID,
		Title:          gb.Title,
		Description:    gb.Description,
		CoverImage:     gb.CoverImage,
		Status:         int(gb.Status),
		ExchangeRate:   gb.ExchangeRate,
		RoundingMethod: rm,
		RoundingDigit:  rd,
		SourceCurrency: gb.SourceCurrency,
		CreatorID:      gb.CreatorID,
		Managers:       managers,

		ShippingConfigs: gb.ShippingConfigs,
		Products:        products,
		CreatedAt:       gb.CreatedAt,
		Deadline:        gb.Deadline,
	}
}

func (p *Product) ToDomain() *groupbuy.Product {
	var specs []*groupbuy.ProductSpec
	for _, s := range p.Specs {
		specs = append(specs, &groupbuy.ProductSpec{
			ID: s.ID, ProductID: s.ProductID, Name: s.Name,
		})
	}
	return &groupbuy.Product{
		ID:            p.ID,
		GroupBuyID:    p.GroupBuyID,
		Name:          p.Name,
		Description:   p.Description,
		ImageURL:      p.ImageURL,
		PriceOriginal: p.PriceOriginal,
		ExchangeRate:  p.ExchangeRate,
		PriceFinal:    p.PriceFinal,
		MaxQuantity:   p.MaxQuantity,
		Rounding:      &groupbuy.RoundingConfig{Method: groupbuy.RoundingMethod(p.RoundingMethod), Digit: p.RoundingDigit},
		Specs:         specs,
	}
}

func FromDomainProduct(p *groupbuy.Product) *Product {
	var specs []*ProductSpec
	for _, s := range p.Specs {
		specs = append(specs, &ProductSpec{ID: s.ID, ProductID: s.ProductID, Name: s.Name})
	}
	// Handle nil rounding
	rm, rd := 0, 0
	if p.Rounding != nil {
		rm, rd = int(p.Rounding.Method), p.Rounding.Digit
	}
	return &Product{
		ID:             p.ID,
		GroupBuyID:     p.GroupBuyID,
		Name:           p.Name,
		Description:    p.Description,
		ImageURL:       p.ImageURL,
		PriceOriginal:  p.PriceOriginal,
		ExchangeRate:   p.ExchangeRate,
		PriceFinal:     p.PriceFinal,
		MaxQuantity:    p.MaxQuantity,
		RoundingMethod: rm,
		RoundingDigit:  rd,
		Specs:          specs,
	}
}

// Order Mappers
func (o *Order) ToDomain() *groupbuy.Order {
	var items []*groupbuy.OrderItem
	for _, i := range o.Items {
		items = append(items, &groupbuy.OrderItem{
			ID: i.ID, OrderID: i.OrderID, ProductID: i.ProductID, SpecID: i.SpecID, Quantity: i.Quantity, Status: groupbuy.OrderItemStatus(i.Status),
			ProductName: i.ProductName, SpecName: i.SpecName, Price: i.Price,
		})
	}
	return &groupbuy.Order{
		ID:              o.ID,
		GroupBuyID:      o.GroupBuyID,
		UserID:          o.UserID,
		TotalAmount:     o.TotalAmount,
		PaymentStatus:   groupbuy.PaymentStatus(o.PaymentStatus),
		ContactInfo:     o.ContactInfo,
		ShippingAddress: o.ShippingAddress,
		PaymentInfo: &groupbuy.PaymentInfo{
			Method:       o.PaymentMethod,
			AccountLast5: o.PaymentAccountLast5,
			PaidAt:       o.PaidAt,
			Amount:       o.PaymentAmount,
		},
		Items:            items,
		ShippingMethodID: o.ShippingMethodID,
		ShippingFee:      o.ShippingFee,
		Note:             o.Note,
		CreatedAt:        o.CreatedAt,
	}
}

func FromDomainOrder(o *groupbuy.Order) *Order {
	var items []*OrderItem
	for _, i := range o.Items {
		items = append(items, &OrderItem{
			ID: i.ID, OrderID: i.OrderID, ProductID: i.ProductID, SpecID: i.SpecID, Quantity: i.Quantity, Status: int(i.Status),
			ProductName: i.ProductName, SpecName: i.SpecName, Price: i.Price,
		})
	}

	var method, account string
	var paidAt *time.Time
	var amount int64
	if o.PaymentInfo != nil {
		method = o.PaymentInfo.Method
		account = o.PaymentInfo.AccountLast5
		paidAt = o.PaymentInfo.PaidAt
		amount = o.PaymentInfo.Amount
	}

	return &Order{
		ID:                  o.ID,
		GroupBuyID:          o.GroupBuyID,
		UserID:              o.UserID,
		TotalAmount:         o.TotalAmount,
		PaymentStatus:       int(o.PaymentStatus),
		ContactInfo:         o.ContactInfo,
		ShippingAddress:     o.ShippingAddress,
		PaymentMethod:       method,
		PaymentAccountLast5: account,
		PaidAt:              paidAt,
		PaymentAmount:       amount,
		Items:               items,
		ShippingMethodID:    o.ShippingMethodID,
		ShippingFee:         o.ShippingFee,
		Note:                o.Note,
		CreatedAt:           o.CreatedAt,
	}
}

// Helper for User
func (u *User) ToDomainValid() *user.User {
	if u == nil {
		return nil
	}
	return u.ToDomain()
}
