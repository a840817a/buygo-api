package handler

import (
	"context"
	"errors"
	"strconv"

	"connectrpc.com/connect"

	v1 "github.com/hatsubosi/buygo-api/api/v1"
	"github.com/hatsubosi/buygo-api/api/v1/buygov1connect"
	"github.com/hatsubosi/buygo-api/internal/domain/auth"
	"github.com/hatsubosi/buygo-api/internal/domain/user"
	"github.com/hatsubosi/buygo-api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Ensure implementation
var _ buygov1connect.AuthServiceHandler = (*AuthHandler)(nil)

func (h *AuthHandler) Login(ctx context.Context, req *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {
	token := req.Msg.IdToken
	if token == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("missing id_token"))
	}

	accessToken, u, err := h.authService.LoginOrRegister(ctx, token)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.LoginResponse{
		AccessToken: accessToken,
		User:        toProtoUser(u),
	}), nil
}

func (h *AuthHandler) GetMe(ctx context.Context, req *connect.Request[v1.GetMeRequest]) (*connect.Response[v1.GetMeResponse], error) {
	userID, _, ok := auth.FromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	u, err := h.authService.GetMe(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.GetMeResponse{
		User: toProtoUser(u),
	}), nil
}

func (h *AuthHandler) ListUsers(ctx context.Context, req *connect.Request[v1.ListUsersRequest]) (*connect.Response[v1.ListUsersResponse], error) {
	_, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}
	if role != int(user.UserRoleSysAdmin) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("admin access required"))
	}

	limit := int(req.Msg.PageSize)
	limit = normalizePageSize(limit)

	offset, err := decodePageToken(req.Msg.PageToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid page_token"))
	}

	users, err := h.authService.ListUsers(ctx, limit+1, offset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	nextPageToken := ""
	if len(users) > limit {
		users = users[:limit]
		nextPageToken = encodePageToken(offset + limit)
	}

	var protoUsers []*v1.User
	for _, u := range users {
		protoUsers = append(protoUsers, toProtoUser(u))
	}

	return connect.NewResponse(&v1.ListUsersResponse{
		Users:         protoUsers,
		NextPageToken: nextPageToken,
	}), nil
}

func decodePageToken(token string) (int, error) {
	if token == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(token)
	if err != nil || offset < 0 {
		return 0, errors.New("invalid page token")
	}
	return offset, nil
}

func encodePageToken(offset int) string {
	if offset <= 0 {
		return ""
	}
	return strconv.Itoa(offset)
}

func normalizePageSize(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func (h *AuthHandler) UpdateUserRole(ctx context.Context, req *connect.Request[v1.UpdateUserRoleRequest]) (*connect.Response[v1.UpdateUserRoleResponse], error) {
	_, role, ok := auth.FromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}
	if role != int(user.UserRoleSysAdmin) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("admin access required"))
	}

	targetUserID := req.Msg.UserId
	newRole := user.UserRole(req.Msg.Role) // Proto enum to Domain enum mapping

	// Minimal validation
	if newRole == user.UserRoleUnspecified {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid role"))
	}

	u, err := h.authService.UpdateUserRole(ctx, targetUserID, newRole)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.UpdateUserRoleResponse{
		User: toProtoUser(u),
	}), nil
}

func (h *AuthHandler) ListAssignableManagers(ctx context.Context, req *connect.Request[v1.ListAssignableManagersRequest]) (*connect.Response[v1.ListAssignableManagersResponse], error) {
	managers, err := h.authService.ListAssignableManagers(ctx, req.Msg.Query)
	if err != nil {
		return nil, mapError(err)
	}

	var protoManagers []*v1.User
	for _, u := range managers {
		protoManagers = append(protoManagers, toProtoUser(u))
	}

	return connect.NewResponse(&v1.ListAssignableManagersResponse{
		Managers: protoManagers,
	}), nil
}

func toProtoUser(u *user.User) *v1.User {
	if u == nil {
		return nil
	}
	// Map Domain Role to Proto Role
	var role v1.UserRole
	switch u.Role {
	case user.UserRoleUser:
		role = v1.UserRole_USER_ROLE_USER
	case user.UserRoleCreator:
		role = v1.UserRole_USER_ROLE_CREATOR
	case user.UserRoleSysAdmin:
		role = v1.UserRole_USER_ROLE_SYS_ADMIN
	default:
		role = v1.UserRole_USER_ROLE_UNSPECIFIED
	}

	return &v1.User{
		Id:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		PhotoUrl: u.PhotoURL,
		Role:     role,
	}
}
