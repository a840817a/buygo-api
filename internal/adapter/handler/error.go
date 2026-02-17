package handler

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/hatsubosi/buygo-api/internal/service"
)

func mapError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	// Authorization
	case errors.Is(err, service.ErrUnauthorized):
		return connect.NewError(connect.CodeUnauthenticated, err)
	case errors.Is(err, service.ErrPermissionDenied):
		return connect.NewError(connect.CodePermissionDenied, err)

	// Not found
	case errors.Is(err, service.ErrNotFound),
		errors.Is(err, service.ErrProductNotFound),
		errors.Is(err, service.ErrSpecNotFound):
		return connect.NewError(connect.CodeNotFound, err)

	// Validation
	case errors.Is(err, service.ErrInvalidQuantity),
		errors.Is(err, service.ErrInvalidShippingMethod),
		errors.Is(err, service.ErrInvalidEventItemID),
		errors.Is(err, service.ErrInvalidStatus),
		errors.Is(err, service.ErrQuantityLimitExceeded):
		return connect.NewError(connect.CodeInvalidArgument, err)

	// Business logic / precondition
	case errors.Is(err, service.ErrNotActive),
		errors.Is(err, service.ErrDeadlinePassed),
		errors.Is(err, service.ErrAlreadyRegistered),
		errors.Is(err, service.ErrPaymentConfirmed),
		errors.Is(err, service.ErrItemsProcessed),
		errors.Is(err, service.ErrModificationDenied):
		return connect.NewError(connect.CodeFailedPrecondition, err)

	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
