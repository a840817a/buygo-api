package handler

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/buygo/buygo-api/internal/service"
)

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, service.ErrPermissionDenied) {
		return connect.NewError(connect.CodePermissionDenied, err)
	}
	// For now, default to Internal for most errors
	// We can add more mappings here (e.g. InvalidArgument)
	return connect.NewError(connect.CodeInternal, err)
}

func mapErrorWithNotFound(err error) error {
	if err == nil {
		return nil
	}
	// Special helper where we assume error likely means not found or mapped
	// This maintains current behavior of Get* methods returning NotFound on error
	// TODO: Proper NotFound error check
	return connect.NewError(connect.CodeNotFound, err)
}
