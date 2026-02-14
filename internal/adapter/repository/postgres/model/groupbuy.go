package model

import (
	"time"

	"github.com/buygo/buygo-api/internal/domain/groupbuy"
	"github.com/buygo/buygo-api/internal/domain/user"
)

type Project struct {
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

	ShippingConfigs []*project.ShippingConfig `gorm:"serializer:json"` // Store as JSON
	Products        []*Product                `gorm:"foreignKey:ProjectID"`
	CreatedAt       time.Time
	Deadline        *time.Time
}

type Product struct {
	ID             string `gorm:"primaryKey"`
	ProjectID      string
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
	ProjectID           string
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

func (p *Project) ToDomain() *project.Project {
	var managers []*user.User
	for _, m := range p.Managers {
		managers = append(managers, m.ToDomain())
	}
	var managerIDs []string
	for _, m := range p.Managers {
		managerIDs = append(managerIDs, m.ID)
	}

	// Products mapper ...
	var products []*project.Product
	for _, prod := range p.Products {
		products = append(products, prod.ToDomain())
	}

	return &project.Project{
		ID:           p.ID,
		Title:        p.Title,
		Description:  p.Description,
		CoverImage:   p.CoverImage,
		Status:       project.ProjectStatus(p.Status),
		ExchangeRate: p.ExchangeRate,
		Rounding:     &project.RoundingConfig{Method: p.RoundingMethod, Digit: p.RoundingDigit},
		CreatorID:    p.CreatorID,
		Creator:      p.Creator.ToDomainValid(), // Handle nil
		ManagerIDs:   managerIDs,
		Managers:     managers,

		ShippingConfigs: p.ShippingConfigs,
		Products:        products,
		CreatedAt:       p.CreatedAt,
		Deadline:        p.Deadline,
		SourceCurrency:  p.SourceCurrency,
	}
}

func FromDomainProject(p *project.Project) *Project {
	var managers []*User
	// Note: For create/update, GORM handles association if we provide the struct with ID.
	// We might need to fetch users first or just set IDs references if GORM supports it.
	// Simplest for ManyToMany: Provide full User structs with just ID set is often enough for association update.
	for _, id := range p.ManagerIDs {
		managers = append(managers, &User{ID: id})
	}

	var products []*Product
	for _, prod := range p.Products {
		products = append(products, FromDomainProduct(prod))
	}

	var rm, rd int
	if p.Rounding != nil {
		rm, rd = p.Rounding.Method, p.Rounding.Digit
	}

	return &Project{
		ID:             p.ID,
		Title:          p.Title,
		Description:    p.Description,
		CoverImage:     p.CoverImage,
		Status:         int(p.Status),
		ExchangeRate:   p.ExchangeRate,
		RoundingMethod: rm,
		RoundingDigit:  rd,
		SourceCurrency: p.SourceCurrency,
		CreatorID:      p.CreatorID,
		Managers:       managers,

		ShippingConfigs: p.ShippingConfigs,
		Products:        products,
		CreatedAt:       p.CreatedAt,
		Deadline:        p.Deadline,
	}
}

func (p *Product) ToDomain() *project.Product {
	var specs []*project.ProductSpec
	for _, s := range p.Specs {
		specs = append(specs, &project.ProductSpec{
			ID: s.ID, ProductID: s.ProductID, Name: s.Name,
		})
	}
	return &project.Product{
		ID:            p.ID,
		ProjectID:     p.ProjectID,
		Name:          p.Name,
		Description:   p.Description,
		ImageURL:      p.ImageURL,
		PriceOriginal: p.PriceOriginal,
		ExchangeRate:  p.ExchangeRate,
		PriceFinal:    p.PriceFinal,
		MaxQuantity:   p.MaxQuantity,
		Rounding:      &project.RoundingConfig{Method: p.RoundingMethod, Digit: p.RoundingDigit},
		Specs:         specs,
	}
}

func FromDomainProduct(p *project.Product) *Product {
	var specs []*ProductSpec
	for _, s := range p.Specs {
		specs = append(specs, &ProductSpec{ID: s.ID, ProductID: s.ProductID, Name: s.Name})
	}
	// Handle nil rounding
	rm, rd := 0, 0
	if p.Rounding != nil {
		rm, rd = p.Rounding.Method, p.Rounding.Digit
	}
	return &Product{
		ID:             p.ID,
		ProjectID:      p.ProjectID,
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
func (o *Order) ToDomain() *project.Order {
	var items []*project.OrderItem
	for _, i := range o.Items {
		items = append(items, &project.OrderItem{
			ID: i.ID, OrderID: i.OrderID, ProductID: i.ProductID, SpecID: i.SpecID, Quantity: i.Quantity, Status: i.Status,
			ProductName: i.ProductName, SpecName: i.SpecName, Price: i.Price,
		})
	}
	return &project.Order{
		ID:              o.ID,
		ProjectID:       o.ProjectID,
		UserID:          o.UserID,
		TotalAmount:     o.TotalAmount,
		PaymentStatus:   o.PaymentStatus,
		ContactInfo:     o.ContactInfo,
		ShippingAddress: o.ShippingAddress,
		PaymentInfo: &project.PaymentInfo{
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

func FromDomainOrder(o *project.Order) *Order {
	var items []*OrderItem
	for _, i := range o.Items {
		items = append(items, &OrderItem{
			ID: i.ID, OrderID: i.OrderID, ProductID: i.ProductID, SpecID: i.SpecID, Quantity: i.Quantity, Status: i.Status,
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
		ProjectID:           o.ProjectID,
		UserID:              o.UserID,
		TotalAmount:         o.TotalAmount,
		PaymentStatus:       o.PaymentStatus,
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
