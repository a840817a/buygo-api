package groupbuy

import (
	"time"

	"github.com/google/uuid"
)

type PriceTemplate struct {
	ID             string
	Name           string
	SourceCurrency string
	ExchangeRate   float64
	Rounding       *RoundingConfig
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
