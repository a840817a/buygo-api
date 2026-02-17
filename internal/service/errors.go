package service

import "errors"

// Authorization
var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrUnauthorized     = errors.New("unauthorized")
)

// Resource not found
var (
	ErrNotFound        = errors.New("resource not found")
	ErrProductNotFound = errors.New("product not found")
	ErrSpecNotFound    = errors.New("spec not found")
)

// Validation
var (
	ErrInvalidQuantity       = errors.New("invalid quantity")
	ErrInvalidShippingMethod = errors.New("invalid shipping method")
	ErrInvalidEventItemID    = errors.New("invalid event item id")
	ErrInvalidStatus         = errors.New("invalid target status")
	ErrQuantityLimitExceeded = errors.New("quantity limit exceeded")
)

// Business logic / precondition failures
var (
	ErrNotActive          = errors.New("not active")
	ErrDeadlinePassed     = errors.New("deadline passed")
	ErrAlreadyRegistered  = errors.New("already registered")
	ErrPaymentConfirmed   = errors.New("payment already confirmed")
	ErrItemsProcessed     = errors.New("items already processed")
	ErrModificationDenied = errors.New("modification not allowed")
)
