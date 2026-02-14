package project

import (
	"time"

	"github.com/google/uuid"
)

type PriceTemplate struct {
	ID             string          `json:"id" gorm:"primaryKey"`
	Name           string          `json:"name"`
	SourceCurrency string          `json:"source_currency"`
	ExchangeRate   float64         `json:"exchange_rate"`
	Rounding       *RoundingConfig `json:"rounding" gorm:"serializer:json"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func NewPriceTemplate(name, sourceCurrency string, rate float64, rounding *RoundingConfig) *PriceTemplate {
	return &PriceTemplate{
		ID:             uuid.New().String(),
		Name:           name,
		SourceCurrency: sourceCurrency,
		ExchangeRate:   rate,
		Rounding:       rounding,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}
