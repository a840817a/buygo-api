package interceptor

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/buygo/buygo-api/internal/domain/auth"
	"github.com/buygo/buygo-api/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Mock TokenManager
type mockTokenManager struct {
	claims *auth.Claims
	err    error
}

func (m *mockTokenManager) GenerateToken(u *user.User) (string, error) {
	return "mock-token", nil
}

func (m *mockTokenManager) ParseToken(token string) (*auth.Claims, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.claims, nil
}

// createTestServer creates a connect unary handler wrapped with the auth interceptor.
// The procedure name determines if it's treated as public or private.
// The handler captures the context so tests can verify user injection.
func createTestServer(
	tm auth.TokenManager,
	procedure string,
	ctxChan chan context.Context,
) *httptest.Server {
	ai := NewAuthInterceptor(tm)

	handler := connect.NewUnaryHandler(
		procedure,
		func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
			ctxChan <- ctx
			return connect.NewResponse(&emptypb.Empty{}), nil
		},
		connect.WithInterceptors(ai.NewUnaryInterceptor()),
	)

	mux := http.NewServeMux()
	mux.Handle(procedure, handler)
	return httptest.NewServer(mux)
}

// doRequest makes a POST request to the test server with optional auth token.
func doRequest(t *testing.T, serverURL, procedure, token string) *http.Response {
	t.Helper()
	// Connect-RPC uses POST with content-type application/proto or application/json
	body := strings.NewReader("{}")
	req, err := http.NewRequest("POST", serverURL+procedure, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func TestAuthInterceptor_PublicEndpoint_NoToken(t *testing.T) {
	tm := &mockTokenManager{err: errors.New("no token")}
	procedure := "/buygo.v1.BuygoService/ListGroupBuys"
	ctxChan := make(chan context.Context, 1)
	server := createTestServer(tm, procedure, ctxChan)
	defer server.Close()

	resp := doRequest(t, server.URL, procedure, "")
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	ctx := <-ctxChan
	_, _, ok := auth.FromContext(ctx)
	assert.False(t, ok, "No user context should be set without token")
}

func TestAuthInterceptor_PublicEndpoint_WithValidToken(t *testing.T) {
	tm := &mockTokenManager{
		claims: &auth.Claims{UserID: "user-1", Role: user.UserRoleUser},
	}
	procedure := "/buygo.v1.BuygoService/GetGroupBuy"
	ctxChan := make(chan context.Context, 1)
	server := createTestServer(tm, procedure, ctxChan)
	defer server.Close()

	resp := doRequest(t, server.URL, procedure, "valid-token")
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	ctx := <-ctxChan
	userID, role, ok := auth.FromContext(ctx)
	assert.True(t, ok, "User context should be injected with valid token")
	assert.Equal(t, "user-1", userID)
	assert.Equal(t, int(user.UserRoleUser), role)
}

func TestAuthInterceptor_PublicEndpoint_InvalidToken(t *testing.T) {
	tm := &mockTokenManager{err: errors.New("invalid")}
	procedure := "/buygo.v1.BuygoService/ListEvents"
	ctxChan := make(chan context.Context, 1)
	server := createTestServer(tm, procedure, ctxChan)
	defer server.Close()

	resp := doRequest(t, server.URL, procedure, "bad-token")
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	ctx := <-ctxChan
	_, _, ok := auth.FromContext(ctx)
	assert.False(t, ok, "No user context with invalid token on public endpoint")
}

func TestAuthInterceptor_PrivateEndpoint_NoToken(t *testing.T) {
	tm := &mockTokenManager{}
	procedure := "/buygo.v1.BuygoService/CreateGroupBuy"
	ctxChan := make(chan context.Context, 1)
	server := createTestServer(tm, procedure, ctxChan)
	defer server.Close()

	resp := doRequest(t, server.URL, procedure, "")
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	// Connect returns 401 for Unauthenticated
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Empty(t, ctxChan, "Handler should not be called")
}

func TestAuthInterceptor_PrivateEndpoint_ValidToken(t *testing.T) {
	tm := &mockTokenManager{
		claims: &auth.Claims{UserID: "creator-1", Role: user.UserRoleCreator},
	}
	procedure := "/buygo.v1.BuygoService/CreateGroupBuy"
	ctxChan := make(chan context.Context, 1)
	server := createTestServer(tm, procedure, ctxChan)
	defer server.Close()

	resp := doRequest(t, server.URL, procedure, "valid-token")
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	ctx := <-ctxChan
	userID, role, ok := auth.FromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "creator-1", userID)
	assert.Equal(t, int(user.UserRoleCreator), role)
}

func TestAuthInterceptor_PrivateEndpoint_InvalidToken(t *testing.T) {
	tm := &mockTokenManager{err: errors.New("expired")}
	procedure := "/buygo.v1.BuygoService/UpdateOrder"
	ctxChan := make(chan context.Context, 1)
	server := createTestServer(tm, procedure, ctxChan)
	defer server.Close()

	resp := doRequest(t, server.URL, procedure, "expired-token")
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Empty(t, ctxChan, "Handler should not be called")
}

func TestAuthInterceptor_LoginEndpoint_NoToken(t *testing.T) {
	tm := &mockTokenManager{}
	procedure := "/buygo.v1.AuthService/Login"
	ctxChan := make(chan context.Context, 1)
	server := createTestServer(tm, procedure, ctxChan)
	defer server.Close()

	resp := doRequest(t, server.URL, procedure, "")
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, ctxChan, 1, "Handler should be called for Login")
}
