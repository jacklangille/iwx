package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Target struct {
	Name        string
	DatabaseURL string
	Dir         string
}

func Run(ctx context.Context, target Target) error {
	if strings.TrimSpace(target.DatabaseURL) == "" {
		return fmt.Errorf("database url is required for target %s", target.Name)
	}

	db, err := sql.Open("pgx", target.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	unlock, err := advisoryLock(ctx, db, target.Name)
	if err != nil {
		return err
	}
	defer unlock()

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return err
	}

	files, err := migrationFiles(target.Dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		applied, err := migrationApplied(ctx, db, file.Name())
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		body, err := os.ReadFile(filepath.Join(target.Dir, file.Name()))
		if err != nil {
			return err
		}

		if err := applyMigration(ctx, db, file.Name(), string(body)); err != nil {
			return err
		}
	}

	return nil
}

func migrationFiles(dir string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]os.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, entry)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	return files, nil
}

func migrationApplied(ctx context.Context, db *sql.DB, filename string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)`, filename).Scan(&exists)
	return exists, err
}

func applyMigration(ctx context.Context, db *sql.DB, filename, sqlText string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, sqlText); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("apply migration %s: %w", filename, err)
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (filename, applied_at) VALUES ($1, $2)`, filename, time.Now().UTC()); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration %s: %w", filename, err)
	}

	return tx.Commit()
}

func advisoryLock(ctx context.Context, db *sql.DB, targetName string) (func(), error) {
	if _, err := db.ExecContext(ctx, `SELECT pg_advisory_lock(hashtext($1))`, "migrations:"+targetName); err != nil {
		return nil, err
	}

	return func() {
		_, _ = db.ExecContext(context.Background(), `SELECT pg_advisory_unlock(hashtext($1))`, "migrations:"+targetName)
	}, nil
}
