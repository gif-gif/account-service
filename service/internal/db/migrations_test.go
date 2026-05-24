package db

import (
	"context"
	"os"
	"testing"

	"account-service/service/internal/testutil"
)

func TestInitMigrationCreatesRequiredTables(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool := testutil.OpenTestDB(t, ctx, databaseURL)
	testutil.ResetSchema(t, ctx, pool)

	if err := ApplyMigrations(ctx, pool); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	requiredTables := []string{
		"accounts",
		"account_leases",
		"api_callers",
		"admin_users",
		"admin_sessions",
		"audit_logs",
	}
	for _, table := range requiredTables {
		var exists bool
		err := pool.QueryRow(ctx, `
			select exists (
				select 1
				from information_schema.tables
				where table_schema = 'public'
					and table_name = $1
			)
		`, table).Scan(&exists)
		if err != nil {
			t.Fatalf("query table %s: %v", table, err)
		}
		if !exists {
			t.Fatalf("expected table %s to exist", table)
		}
	}
}
