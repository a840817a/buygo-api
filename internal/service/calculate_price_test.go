package service

import (
	"testing"

	"github.com/buygo/buygo-api/internal/domain/project"
	"github.com/stretchr/testify/assert"
)

func TestCalculateFinalPrice(t *testing.T) {
	svc := &ProjectService{}

	tests := []struct {
		name     string
		original int64
		rate     float64
		rounding *project.RoundingConfig
		expected int64
	}{
		// Zero cases
		{"zero original", 0, 0.25, nil, 0},
		{"zero rate", 100, 0, nil, 0},

		// Nil rounding → default Floor/Ones
		{"nil rounding defaults to floor ones", 100, 0.23, nil, 23},

		// Floor rounding (method=1)
		{"floor ones 100*0.23=23", 100, 0.23, &project.RoundingConfig{Method: 1, Digit: 0}, 23},
		{"floor ones 199*0.23=45.77→45", 199, 0.23, &project.RoundingConfig{Method: 1, Digit: 0}, 45},
		{"floor tens 199*0.23=45.77→40", 199, 0.23, &project.RoundingConfig{Method: 1, Digit: 1}, 40},
		{"floor hundreds 5000*0.23=1150→1100", 5000, 0.23, &project.RoundingConfig{Method: 1, Digit: 2}, 1100},

		// Ceil rounding (method=2)
		{"ceil ones 100*0.23=23→23", 100, 0.23, &project.RoundingConfig{Method: 2, Digit: 0}, 23},
		{"ceil ones 199*0.23=45.77→46", 199, 0.23, &project.RoundingConfig{Method: 2, Digit: 0}, 46},
		{"ceil tens 199*0.23=45.77→50", 199, 0.23, &project.RoundingConfig{Method: 2, Digit: 1}, 50},

		// Round rounding (method=3)
		{"round ones 199*0.23=45.77→46", 199, 0.23, &project.RoundingConfig{Method: 3, Digit: 0}, 46},
		{"round ones 100*0.22=22→22", 100, 0.22, &project.RoundingConfig{Method: 3, Digit: 0}, 22},
		{"round tens 199*0.23=45.77→50", 199, 0.23, &project.RoundingConfig{Method: 3, Digit: 1}, 50},
		{"round tens 120*0.23=27.6→30", 120, 0.23, &project.RoundingConfig{Method: 3, Digit: 1}, 30},
		{"round tens 110*0.23=25.3→30", 110, 0.23, &project.RoundingConfig{Method: 3, Digit: 1}, 30},

		// Large values
		{"large value floor", 100000, 0.25, &project.RoundingConfig{Method: 1, Digit: 0}, 25000},
		{"large value ceil hundreds", 100001, 0.25, &project.RoundingConfig{Method: 2, Digit: 2}, 25100},

		// Rate > 1 (e.g. JPY → TWD where TWD is more expensive)
		{"rate greater than 1", 100, 4.5, &project.RoundingConfig{Method: 1, Digit: 0}, 450},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CalculateFinalPrice(tt.original, tt.rate, tt.rounding)
			assert.Equal(t, tt.expected, result)
		})
	}
}
