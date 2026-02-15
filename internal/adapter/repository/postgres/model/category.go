package model

import (
	"github.com/buygo/buygo-api/internal/domain/groupbuy"
	"github.com/lib/pq"
)

type Category struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	SpecNames pq.StringArray `gorm:"type:text[]"`
}

func (c *Category) ToDomain() *groupbuy.Category {
	return &groupbuy.Category{
		ID:        c.ID,
		Name:      c.Name,
		SpecNames: c.SpecNames,
	}
}

func FromDomainCategory(c *groupbuy.Category) *Category {
	return &Category{
		ID:        c.ID,
		Name:      c.Name,
		SpecNames: c.SpecNames,
	}
}
