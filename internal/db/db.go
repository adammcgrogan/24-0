package db

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaDDL string

// ErrNotConnected is returned by DB functions when the pool has not been
// successfully initialised (e.g. DATABASE_URL was not set at startup).
var ErrNotConnected = errors.New("database not connected")

var pool *pgxpool.Pool

func Connect(ctx context.Context) error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL not set")
	}
	p, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("pgxpool.New: %w", err)
	}
	pool = p
	if _, err := pool.Exec(ctx, schemaDDL); err != nil {
		return fmt.Errorf("schema migration: %w", err)
	}
	return nil
}

func Pool() *pgxpool.Pool {
	return pool
}

func Close() {
	if pool != nil {
		pool.Close()
	}
}

func checkPool() error {
	if pool == nil {
		return ErrNotConnected
	}
	return nil
}
