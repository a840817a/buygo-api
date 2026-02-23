package handler

import (
	"fmt"

	"connectrpc.com/connect"
)

// validateRequired returns an InvalidArgument error if field is empty.
func validateRequired(field, name string) error {
	if field == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("%s is required", name))
	}
	return nil
}

// validateMaxLength returns an InvalidArgument error if field exceeds max runes.
func validateMaxLength(field, name string, max int) error {
	if len([]rune(field)) > max {
		return connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("%s must not exceed %d characters", name, max))
	}
	return nil
}

// validatePositiveInt64 returns an InvalidArgument error if val <= 0.
func validatePositiveInt64(val int64, name string) error {
	if val <= 0 {
		return connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("%s must be greater than zero", name))
	}
	return nil
}

// validatePositiveFloat64 returns an InvalidArgument error if val <= 0.
func validatePositiveFloat64(val float64, name string) error {
	if val <= 0 {
		return connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("%s must be greater than zero", name))
	}
	return nil
}
