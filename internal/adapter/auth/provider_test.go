package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockProvider_VerifyToken_Success(t *testing.T) {
	p := NewMockProvider()
	info, err := p.VerifyToken(context.Background(), "any-valid-token")
	require.NoError(t, err)
	assert.Equal(t, "mock-user-123", info.UID)
	assert.Equal(t, "test@example.com", info.Email)
	assert.Equal(t, "Test User", info.Name)
	assert.Equal(t, "https://example.com/avatar.jpg", info.AvatarURL)
}

func TestMockProvider_VerifyToken_Invalid(t *testing.T) {
	p := NewMockProvider()
	_, err := p.VerifyToken(context.Background(), "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestMockProvider_VerifyToken_EmptyToken(t *testing.T) {
	p := NewMockProvider()
	// Empty token is not "invalid", so it should succeed with mock data
	info, err := p.VerifyToken(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "mock-user-123", info.UID)
}

func TestMockProvider_ImplementsProvider(t *testing.T) {
	var _ Provider = (*MockProvider)(nil)
}

func TestFirebaseProvider_MockMode_DefaultToken(t *testing.T) {
	fp := &FirebaseProvider{MockMode: true}
	info, err := fp.VerifyToken(context.Background(), "some-token")
	require.NoError(t, err)
	assert.Equal(t, "test-user-id", info.UID)
	assert.Equal(t, "test@example.com", info.Email)
	assert.Equal(t, "Test User", info.Name)
	assert.Equal(t, "https://via.placeholder.com/150", info.PhotoURL)
}

func TestFirebaseProvider_MockMode_MockTokenPrefix(t *testing.T) {
	fp := &FirebaseProvider{MockMode: true}
	info, err := fp.VerifyToken(context.Background(), "mock-token-alice")
	require.NoError(t, err)
	assert.Equal(t, "alice", info.UID)
	assert.Equal(t, "User alice", info.Name)
	assert.Equal(t, "alice@example.com", info.Email)
}

func TestFirebaseProvider_MockMode_ShortMockToken(t *testing.T) {
	fp := &FirebaseProvider{MockMode: true}
	// Token is "mock-token-" (11 chars) with no uid suffix — too short for prefix check
	info, err := fp.VerifyToken(context.Background(), "mock-token-")
	require.NoError(t, err)
	// len("mock-token-") == 11, not > 11, so falls through to default
	assert.Equal(t, "test-user-id", info.UID)
}

func TestFirebaseProvider_MockMode_ExactPrefix(t *testing.T) {
	fp := &FirebaseProvider{MockMode: true}
	// Token "mock-token-x" has len 12, > 11 so prefix matches
	info, err := fp.VerifyToken(context.Background(), "mock-token-x")
	require.NoError(t, err)
	assert.Equal(t, "x", info.UID)
	assert.Equal(t, "User x", info.Name)
	assert.Equal(t, "x@example.com", info.Email)
}

func TestFirebaseProvider_MockMode_EmptyToken(t *testing.T) {
	fp := &FirebaseProvider{MockMode: true}
	info, err := fp.VerifyToken(context.Background(), "")
	require.NoError(t, err)
	// Empty token doesn't match prefix, gets default values
	assert.Equal(t, "test-user-id", info.UID)
}
