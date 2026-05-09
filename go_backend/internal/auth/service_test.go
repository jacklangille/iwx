package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"iwx/go_backend/internal/config"
)

type stubUserStore struct {
	user User
	err  error
}

func (s stubUserStore) FindUserByUsername(_ context.Context, username string) (User, error) {
	if s.err != nil {
		return User{}, s.err
	}

	if s.user.Username != username {
		return User{}, ErrUserNotFound
	}

	return s.user, nil
}

func (s stubUserStore) CreateUser(_ context.Context, username, passwordHash string) (User, error) {
	if s.err != nil {
		return User{}, s.err
	}

	return User{
		ID:           2,
		Username:     username,
		PasswordHash: passwordHash,
		Active:       true,
	}, nil
}

func TestAuthenticateReturnsJWT(t *testing.T) {
	passwordHash, err := HashPassword("secret-pass")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	service, err := NewService(stubUserStore{
		user: User{
			ID:           1,
			Username:     "alice",
			PasswordHash: passwordHash,
			Active:       true,
		},
	}, config.Config{
		AuthJWTSecret: "jwt-secret",
		AuthJWTIssuer: "iwx-test",
		AuthJWTTTL:    600,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	service.now = func() time.Time {
		return time.Unix(1_700_000_000, 0).UTC()
	}

	token, claims, err := service.Authenticate(context.Background(), "alice", "secret-pass")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if token == "" {
		t.Fatal("expected token to be returned")
	}

	if claims.Subject != "alice" {
		t.Fatalf("expected subject alice, got %q", claims.Subject)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 jwt segments, got %d", len(parts))
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}

	var decoded Claims
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Subject != "alice" {
		t.Fatalf("expected payload subject alice, got %q", decoded.Subject)
	}

	if decoded.ExpiresAt-decoded.IssuedAt != 600 {
		t.Fatalf("expected ttl 600, got %d", decoded.ExpiresAt-decoded.IssuedAt)
	}
}

func TestAuthenticateRejectsInvalidCredentials(t *testing.T) {
	passwordHash, err := HashPassword("secret-pass")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	service, err := NewService(stubUserStore{
		user: User{
			ID:           1,
			Username:     "alice",
			PasswordHash: passwordHash,
			Active:       true,
		},
	}, config.Config{
		AuthJWTSecret: "jwt-secret",
		AuthJWTIssuer: "iwx-test",
		AuthJWTTTL:    600,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, _, err = service.Authenticate(context.Background(), "alice", "wrong-pass")
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
