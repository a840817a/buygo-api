package service

import (
	"context"

	"github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

// Permission Helpers

func checkLogin(ctx context.Context) (string, int, error) {
	userID, role, ok := auth.FromContext(ctx)
	if !ok {
		return "", 0, ErrUnauthorized
	}
	return userID, role, nil
}

func requireRole(ctx context.Context, allowed ...user.UserRole) (string, int, error) {
	userID, role, ok := auth.FromContext(ctx)
	if !ok {
		return "", 0, ErrUnauthorized
	}

	for _, r := range allowed {
		if role == int(r) || role == int(user.UserRoleSysAdmin) {
			return userID, role, nil
		}
	}

	return "", 0, ErrPermissionDenied
}

type ManagerChecker interface {
	IsManager(userID string) bool
}

func canManage(role int, userID string, entity ManagerChecker) bool {
	if role == int(user.UserRoleSysAdmin) {
		return true
	}
	return entity.IsManager(userID)
}

func isSysAdmin(role int) bool {
	return role == int(user.UserRoleSysAdmin)
}
