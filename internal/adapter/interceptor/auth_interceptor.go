package interceptor

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"

	"github.com/buygo/buygo-api/internal/domain/auth"
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
			// Skip auth for public endpoints (like Login)
			// Ideally, we check the procedure name.
			// For now, let's allow "Login" specifically.
			procedure := req.Spec().Procedure
			// Public Endpoints
			if strings.Contains(procedure, "Login") ||
				strings.Contains(procedure, "ListGroupBuys") ||
				strings.Contains(procedure, "GetGroupBuy") ||
				strings.Contains(procedure, "ListEvents") ||
				strings.Contains(procedure, "GetEvent") {
				// For List/Get, we still want to try parsing token if present (optional auth)
				// to allow "GetMyOrders" mixed in or personalized views?
				// Actually "GetMy..." are separate.
				// So for strict public views, we can return next(ctx, req) immediately
				// OR try to parse token to inject user context for "personalized" public view (like "You already liked this").
				// For now, simple Bypass.

				// Optional Auth Logic: If token exists, parse it. If not, proceed as anon.
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

			return next(ctx, req)
		}
	}
}
