package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"iwx/go_backend/internal/config"
)

var (
	ErrNotConfigured      = errors.New("auth service is not configured")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameTaken      = errors.New("username is already taken")
	ErrInvalidSignup      = errors.New("invalid signup")
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Active       bool
}

type UserStore interface {
	FindUserByUsername(ctx context.Context, username string) (User, error)
	CreateUser(ctx context.Context, username, passwordHash string) (User, error)
}

type Service struct {
	users  UserStore
	issuer string
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

type Claims struct {
	Subject   string `json:"sub"`
	UserID    int64  `json:"uid"`
	Issuer    string `json:"iss"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func NewService(users UserStore, cfg config.Config) (*Service, error) {
	if users == nil || strings.TrimSpace(cfg.AuthJWTSecret) == "" {
		return &Service{}, nil
	}

	return &Service{
		users:  users,
		issuer: cfg.AuthJWTIssuer,
		secret: []byte(cfg.AuthJWTSecret),
		ttl:    time.Duration(cfg.AuthJWTTTL) * time.Second,
		now:    time.Now,
	}, nil
}

func (s *Service) Enabled() bool {
	return s != nil && s.users != nil && len(s.secret) > 0
}

func (s *Service) Authenticate(ctx context.Context, username, password string) (string, Claims, error) {
	if !s.Enabled() {
		return "", Claims{}, ErrNotConfigured
	}

	user, err := s.users.FindUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return "", Claims{}, ErrInvalidCredentials
		}
		return "", Claims{}, err
	}

	if !user.Active {
		return "", Claims{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", Claims{}, ErrInvalidCredentials
	}

	now := s.now().UTC()
	claims := Claims{
		Subject:   user.Username,
		UserID:    user.ID,
		Issuer:    s.issuer,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(s.ttl).Unix(),
	}

	token, err := signToken(claims, s.secret)
	if err != nil {
		return "", Claims{}, err
	}

	return token, claims, nil
}

func (s *Service) Signup(ctx context.Context, username, password string) (string, Claims, error) {
	if !s.Enabled() {
		return "", Claims{}, ErrNotConfigured
	}

	username = strings.TrimSpace(username)
	if len(username) < 3 || len(password) < 8 {
		return "", Claims{}, ErrInvalidSignup
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return "", Claims{}, err
	}

	user, err := s.users.CreateUser(ctx, username, passwordHash)
	if err != nil {
		return "", Claims{}, err
	}

	now := s.now().UTC()
	claims := Claims{
		Subject:   user.Username,
		UserID:    user.ID,
		Issuer:    s.issuer,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(s.ttl).Unix(),
	}

	token, err := signToken(claims, s.secret)
	if err != nil {
		return "", Claims{}, err
	}

	return token, claims, nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func signToken(claims Claims, secret []byte) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := encodedHeader + "." + encodedPayload

	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s.%s", signingInput, signature), nil
}

func VerifyToken(token string, cfg config.Config) (Claims, error) {
	if strings.TrimSpace(cfg.AuthJWTSecret) == "" {
		return Claims{}, ErrNotConfigured
	}

	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidCredentials
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(cfg.AuthJWTSecret))
	_, _ = mac.Write([]byte(signingInput))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return Claims{}, ErrInvalidCredentials
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidCredentials
	}

	var claims Claims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return Claims{}, ErrInvalidCredentials
	}

	now := time.Now().UTC().Unix()
	if claims.Issuer != cfg.AuthJWTIssuer || claims.ExpiresAt <= now || claims.UserID <= 0 || strings.TrimSpace(claims.Subject) == "" {
		return Claims{}, ErrInvalidCredentials
	}

	return claims, nil
}
