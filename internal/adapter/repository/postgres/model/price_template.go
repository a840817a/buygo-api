package model

import (
	"time"

	"github.com/hatsubosi/buygo-api/internal/domain/groupbuy"
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

func FromDomainPriceTemplate(pt *groupbuy.PriceTemplate) *PriceTemplate {
	if pt == nil {
		return nil
	}
	var rm, rd int
	if pt.Rounding != nil {
		rm = int(pt.Rounding.Method)
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

func (m *PriceTemplate) ToDomain() *groupbuy.PriceTemplate {
	if m == nil {
		return nil
	}
	return &groupbuy.PriceTemplate{
		ID:             m.ID,
		Name:           m.Name,
		SourceCurrency: m.SourceCurrency,
		ExchangeRate:   m.ExchangeRate,
		Rounding: &groupbuy.RoundingConfig{
			Method: groupbuy.RoundingMethod(m.RoundingMethod),
			Digit:  m.RoundingDigit,
		},
	}
}
