package handler

import (
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/buygo/buygo-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapError_Nil(t *testing.T) {
	assert.Nil(t, mapError(nil))
}

func TestMapError_UnknownError(t *testing.T) {
	err := mapError(errors.New("something unexpected"))
	require.Error(t, err)
	var connectErr *connect.Error
	require.True(t, errors.As(err, &connectErr))
	assert.Equal(t, connect.CodeInternal, connectErr.Code())
}

func TestMapError_AllSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode connect.Code
	}{
		// Authorization
		{"Unauthorized", service.ErrUnauthorized, connect.CodeUnauthenticated},
		{"PermissionDenied", service.ErrPermissionDenied, connect.CodePermissionDenied},

		// Not found
		{"NotFound", service.ErrNotFound, connect.CodeNotFound},
		{"ProductNotFound", service.ErrProductNotFound, connect.CodeNotFound},
		{"SpecNotFound", service.ErrSpecNotFound, connect.CodeNotFound},

		// Validation
		{"InvalidQuantity", service.ErrInvalidQuantity, connect.CodeInvalidArgument},
		{"InvalidShippingMethod", service.ErrInvalidShippingMethod, connect.CodeInvalidArgument},
		{"InvalidEventItemID", service.ErrInvalidEventItemID, connect.CodeInvalidArgument},
		{"InvalidStatus", service.ErrInvalidStatus, connect.CodeInvalidArgument},
		{"QuantityLimitExceeded", service.ErrQuantityLimitExceeded, connect.CodeInvalidArgument},

		// Business logic / precondition
		{"NotActive", service.ErrNotActive, connect.CodeFailedPrecondition},
		{"DeadlinePassed", service.ErrDeadlinePassed, connect.CodeFailedPrecondition},
		{"AlreadyRegistered", service.ErrAlreadyRegistered, connect.CodeFailedPrecondition},
		{"PaymentConfirmed", service.ErrPaymentConfirmed, connect.CodeFailedPrecondition},
		{"ItemsProcessed", service.ErrItemsProcessed, connect.CodeFailedPrecondition},
		{"ModificationDenied", service.ErrModificationDenied, connect.CodeFailedPrecondition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapError(tt.err)
			require.Error(t, err)
			var connectErr *connect.Error
			require.True(t, errors.As(err, &connectErr))
			assert.Equal(t, tt.wantCode, connectErr.Code())
		})
	}
}

func TestMapError_WrappedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode connect.Code
	}{
		{"WrappedNotFound", fmt.Errorf("product xyz: %w", service.ErrProductNotFound), connect.CodeNotFound},
		{"WrappedPaymentConfirmed", fmt.Errorf("order 123: %w", service.ErrPaymentConfirmed), connect.CodeFailedPrecondition},
		{"WrappedInvalidQuantity", fmt.Errorf("item abc: %w", service.ErrInvalidQuantity), connect.CodeInvalidArgument},
		{"WrappedPermissionDenied", fmt.Errorf("user not owner: %w", service.ErrPermissionDenied), connect.CodePermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapError(tt.err)
			require.Error(t, err)
			var connectErr *connect.Error
			require.True(t, errors.As(err, &connectErr))
			assert.Equal(t, tt.wantCode, connectErr.Code())
		})
	}
}
