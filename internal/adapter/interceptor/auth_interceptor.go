package interceptor

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"

	"github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
)

type AuthInterceptor struct {
	tokenManager auth.TokenManager
}

func NewAuthInterceptor(tokenManager auth.TokenManager) *AuthInterceptor {
	return &AuthInterceptor{tokenManager: tokenManager}
}

func (i *AuthInterceptor) NewUnaryInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			procedure := req.Spec().Procedure
			access := procedureAccess(procedure)
			if access == accessPublic {
				// Optional auth on public endpoints.
				token := req.Header().Get("Authorization")
				if token != "" {
					token = strings.TrimPrefix(token, "Bearer ")
					if claims, err := i.tokenManager.ParseToken(token); err == nil {
						ctx = auth.NewContext(ctx, claims.UserID, int(claims.Role))
					}
				}
				return next(ctx, req)
			}

			// For all other endpoints, verify token
			token := req.Header().Get("Authorization")
			if token == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing token"))
			}

			token = strings.TrimPrefix(token, "Bearer ")
			claims, err := i.tokenManager.ParseToken(token)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
			}

			// Inject into context
			ctx = auth.NewContext(ctx, claims.UserID, int(claims.Role))
			if access == accessSysAdmin && claims.Role != user.UserRoleSysAdmin {
				return nil, connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
			}

			return next(ctx, req)
		}
	}
}
