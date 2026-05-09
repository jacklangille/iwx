package postgres

import (
	"context"
	"database/sql"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var dbCache sync.Map

type baseRepository struct {
	db *sql.DB
}

func newBaseRepository(databaseURL string) *baseRepository {
	return &baseRepository{db: openDB(databaseURL)}
}

func openDB(databaseURL string) *sql.DB {
	if existing, ok := dbCache.Load(databaseURL); ok {
		return existing.(*sql.DB)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(5 * time.Minute)

	actual, _ := dbCache.LoadOrStore(databaseURL, db)
	if actual != db {
		_ = db.Close()
	}

	return actual.(*sql.DB)
}

func withTransaction[T any](ctx context.Context, db *sql.DB, fn func(*sql.Tx) (T, error)) (T, error) {
	var zero T

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}

	result, err := fn(tx)
	if err != nil {
		_ = tx.Rollback()
		return zero, err
	}

	if err := tx.Commit(); err != nil {
		return zero, err
	}

	return result, nil
}
