package testutil

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func OpenTestDB(t *testing.T, ctx context.Context, databaseURL string) *pgxpool.Pool {
	t.Helper()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}
	return pool
}

func ResetSchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	if _, err := pool.Exec(ctx, `drop schema public cascade; create schema public;`); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
}
