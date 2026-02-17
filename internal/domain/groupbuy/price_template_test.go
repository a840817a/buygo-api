package groupbuy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPriceTemplate(t *testing.T) {
	before := time.Now()
	pt := NewPriceTemplate("JP Template", "JPY", 0.22, &RoundingConfig{Method: RoundingMethodFloor, Digit: 0})
	after := time.Now()

	assert.NotEmpty(t, pt.ID)
	assert.Equal(t, "JP Template", pt.Name)
	assert.Equal(t, "JPY", pt.SourceCurrency)
	assert.Equal(t, 0.22, pt.ExchangeRate)
	assert.Equal(t, RoundingMethodFloor, pt.Rounding.Method)
	assert.Equal(t, 0, pt.Rounding.Digit)

	// Timestamps should be set to approximately now
	assert.True(t, !pt.CreatedAt.Before(before) && !pt.CreatedAt.After(after))
	assert.True(t, !pt.UpdatedAt.Before(before) && !pt.UpdatedAt.After(after))
}

func TestNewPriceTemplate_NilRounding(t *testing.T) {
	pt := NewPriceTemplate("Simple", "USD", 1.0, nil)
	assert.NotEmpty(t, pt.ID)
	assert.Nil(t, pt.Rounding)
}
