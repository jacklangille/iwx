package authhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"iwx/go_backend/internal/auth"
	"iwx/go_backend/internal/config"
)

func TestHandleAuthLoginReturnsToken(t *testing.T) {
	passwordHash, err := auth.HashPassword("secret-pass")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	authService, err := auth.NewService(stubUserStore{
		user: auth.User{
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

	server := NewServer(config.Config{}, authService)

	body, err := json.Marshal(map[string]string{
		"username": "alice",
		"password": "secret-pass",
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	server.handleAuthLogin(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if response["access_token"] == "" {
		t.Fatalf("expected access_token in response, got %v", response)
	}
}

func TestHandleAuthLoginRejectsBadPassword(t *testing.T) {
	passwordHash, err := auth.HashPassword("secret-pass")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	authService, err := auth.NewService(stubUserStore{
		user: auth.User{
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

	server := NewServer(config.Config{}, authService)

	body, err := json.Marshal(map[string]string{
		"username": "alice",
		"password": "wrong-pass",
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	server.handleAuthLogin(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

type stubUserStore struct {
	user         auth.User
	createUserFn func(context.Context, string, string) (auth.User, error)
}

func (s stubUserStore) FindUserByUsername(_ context.Context, username string) (auth.User, error) {
	if s.user.Username != username {
		return auth.User{}, auth.ErrUserNotFound
	}

	return s.user, nil
}

func (s stubUserStore) CreateUser(ctx context.Context, username, passwordHash string) (auth.User, error) {
	if s.createUserFn != nil {
		return s.createUserFn(ctx, username, passwordHash)
	}

	return auth.User{
		ID:           2,
		Username:     username,
		PasswordHash: passwordHash,
		Active:       true,
	}, nil
}

func TestHandleAuthSignupReturnsToken(t *testing.T) {
	authService, err := auth.NewService(stubUserStore{}, config.Config{
		AuthJWTSecret: "jwt-secret",
		AuthJWTIssuer: "iwx-test",
		AuthJWTTTL:    600,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	server := NewServer(config.Config{}, authService)

	body, err := json.Marshal(map[string]string{
		"username": "new-user",
		"password": "secret-pass",
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	server.handleAuthSignup(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if response["access_token"] == "" {
		t.Fatalf("expected access_token in response, got %v", response)
	}
}

func TestHandleAuthSignupRejectsDuplicateUsername(t *testing.T) {
	authService, err := auth.NewService(stubUserStore{
		createUserFn: func(_ context.Context, _ string, _ string) (auth.User, error) {
			return auth.User{}, auth.ErrUsernameTaken
		},
	}, config.Config{
		AuthJWTSecret: "jwt-secret",
		AuthJWTIssuer: "iwx-test",
		AuthJWTTTL:    600,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	server := NewServer(config.Config{}, authService)

	body, err := json.Marshal(map[string]string{
		"username": "alice",
		"password": "secret-pass",
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	server.handleAuthSignup(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestServerRoutesSignupThroughMux(t *testing.T) {
	authService, err := auth.NewService(stubUserStore{}, config.Config{
		AuthJWTSecret: "jwt-secret",
		AuthJWTIssuer: "iwx-test",
		AuthJWTTTL:    600,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	server := NewServer(config.Config{}, authService)

	body, err := json.Marshal(map[string]string{
		"username": "new-user",
		"password": "secret-pass",
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	server.mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}
