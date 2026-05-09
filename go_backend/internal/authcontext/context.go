package authcontext

import (
	"context"
	"errors"

	"iwx/go_backend/internal/auth"
)

type contextKey string

const claimsKey contextKey = "auth_claims"

var ErrMissingClaims = errors.New("missing auth claims")

func WithClaims(ctx context.Context, claims auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func ClaimsFromContext(ctx context.Context) (auth.Claims, error) {
	value := ctx.Value(claimsKey)
	if value == nil {
		return auth.Claims{}, ErrMissingClaims
	}

	claims, ok := value.(auth.Claims)
	if !ok {
		return auth.Claims{}, ErrMissingClaims
	}

	return claims, nil
}
