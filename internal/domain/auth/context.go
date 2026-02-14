package auth

import "context"

type contextKey string

const (
	userKey contextKey = "user_id"
	roleKey contextKey = "user_role"
)

func NewContext(ctx context.Context, userID string, role int) context.Context {
	ctx = context.WithValue(ctx, userKey, userID)
	ctx = context.WithValue(ctx, roleKey, role)
	return ctx
}

func FromContext(ctx context.Context) (string, int, bool) {
	id, ok1 := ctx.Value(userKey).(string)
	role, ok2 := ctx.Value(roleKey).(int)
	return id, role, ok1 && ok2
}
