package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"iwx/go_backend/internal/auth"
)

type AuthUserRepository struct {
	*baseRepository
}

func NewAuthUserRepository(databaseURL string) *AuthUserRepository {
	return &AuthUserRepository{baseRepository: newBaseRepository(databaseURL)}
}

func (r *AuthUserRepository) FindUserByUsername(ctx context.Context, username string) (auth.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, active
		FROM users
		WHERE username = $1
		LIMIT 1
	`, strings.TrimSpace(username))

	var user auth.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.User{}, auth.ErrUserNotFound
		}
		return auth.User{}, err
	}

	return user, nil
}

func (r *AuthUserRepository) UpsertUser(ctx context.Context, username, passwordHash string, active bool) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (username, password_hash, active, inserted_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (username)
		DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			active = EXCLUDED.active,
			updated_at = NOW()
	`, strings.TrimSpace(username), passwordHash, active)

	return err
}

func (r *AuthUserRepository) CreateUser(ctx context.Context, username, passwordHash string) (auth.User, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO users (username, password_hash, active, inserted_at, updated_at)
		VALUES ($1, $2, TRUE, NOW(), NOW())
		RETURNING id, username, password_hash, active
	`, strings.TrimSpace(username), passwordHash)

	var user auth.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Active)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return auth.User{}, auth.ErrUsernameTaken
		}
		return auth.User{}, err
	}

	return user, nil
}
