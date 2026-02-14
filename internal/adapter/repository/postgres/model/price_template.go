package model

import (
	"time"

	"github.com/buygo/buygo-api/internal/domain/project"
)

type PriceTemplate struct {
	ID             string `gorm:"primaryKey"`
	Name           string
	SourceCurrency string
	ExchangeRate   float64
	RoundingMethod int
	RoundingDigit  int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func FromDomainPriceTemplate(pt *project.PriceTemplate) *PriceTemplate {
	if pt == nil {
		return nil
	}
	var rm, rd int
	if pt.Rounding != nil {
		rm = pt.Rounding.Method
		rd = pt.Rounding.Digit
	}
	return &PriceTemplate{
		ID:             pt.ID,
		Name:           pt.Name,
		SourceCurrency: pt.SourceCurrency,
		ExchangeRate:   pt.ExchangeRate,
		RoundingMethod: rm,
		RoundingDigit:  rd,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func (m *PriceTemplate) ToDomain() *project.PriceTemplate {
	if m == nil {
		return nil
	}
	return &project.PriceTemplate{
		ID:             m.ID,
		Name:           m.Name,
		SourceCurrency: m.SourceCurrency,
		ExchangeRate:   m.ExchangeRate,
		Rounding: &project.RoundingConfig{
			Method: m.RoundingMethod,
			Digit:  m.RoundingDigit,
		},
	}
}
