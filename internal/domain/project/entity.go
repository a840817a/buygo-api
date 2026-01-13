package project

import (
	"context"
	"errors"
	"time"

	"buygo/internal/domain/user"
)

// Project Entity
type Project struct {
	ID           string
	Title        string
	Description  string
	CoverImage   string
	IsActive     bool
	CreatorID    string
	Creator      *user.User
	ManagerIDs   []string
	Managers     []*user.User
	Products     []*Product
	CreatedAt    time.Time
	Deadline     *time.Time
}

type Product struct {
	ID            string
	ProjectID     string
	Name          string
	PriceOriginal int64
	ExchangeRate  float64
	PriceFinal    int64
	Specs         []*ProductSpec
}

type ProductSpec struct {
	ID        string
	ProductID string
	Name      string
}

// Repository Port
type Repository interface {
	Create(ctx context.Context, p *Project) error
	GetByID(ctx context.Context, id string) (*Project, error)
	List(ctx context.Context, limit int, offset int) ([]*Project, error)
	Update(ctx context.Context, p *Project) error
}

// Service Port
type Service interface {
	CreateProject(ctx context.Context, userID string, title string) (*Project, error)
	GetProject(ctx context.Context, id string) (*Project, error)
}
