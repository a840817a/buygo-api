package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"

	"github.com/buygo/buygo-api/internal/domain/auth"
)

type FirebaseProvider struct {
	app *firebase.App
}

func NewFirebaseProvider(credentialsJSON []byte) (*FirebaseProvider, error) {
	opts := []option.ClientOption{}
	if len(credentialsJSON) > 0 {
		opts = append(opts, option.WithCredentialsJSON(credentialsJSON))
	}

	app, err := firebase.NewApp(context.Background(), nil, opts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}

	return &FirebaseProvider{app: app}, nil
}

func (fp *FirebaseProvider) VerifyToken(ctx context.Context, token string) (*auth.TokenInfo, error) {
	// Dev Mode: If app is not initialized, return mock user
	if fp.app == nil {
		// Mock validation for development
		// Check if token starts with "mock-token-"
		uid := "test-user-id"
		name := "Test User"
		email := "test@example.com"

		if len(token) > 11 && token[:11] == "mock-token-" {
			uid = token[11:]
			name = "User " + uid
			email = uid + "@example.com"
		}

		return &auth.TokenInfo{
			UID:      uid,
			Email:    email,
			Name:     name,
			PhotoURL: "https://via.placeholder.com/150",
		}, nil
	}

	client, err := fp.app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting auth client: %w", err)
	}

	t, err := client.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("error verifying token: %w", err)
	}

	claims := t.Claims
	name, _ := claims["name"].(string)
	picture, _ := claims["picture"].(string)
	email, _ := claims["email"].(string)

	return &auth.TokenInfo{
		UID:      t.UID,
		Email:    email,
		Name:     name,
		PhotoURL: picture,
	}, nil
}
