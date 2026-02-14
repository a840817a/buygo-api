package model

import (
	"github.com/buygo/buygo-api/internal/domain/project"
	"github.com/lib/pq"
)

type Category struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	SpecNames pq.StringArray `gorm:"type:text[]"`
}

func (c *Category) ToDomain() *project.Category {
	return &project.Category{
		ID:        c.ID,
		Name:      c.Name,
		SpecNames: c.SpecNames,
	}
}

func FromDomainCategory(c *project.Category) *Category {
	return &Category{
		ID:        c.ID,
		Name:      c.Name,
		SpecNames: c.SpecNames,
	}
}
