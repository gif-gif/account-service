package db

import (
	"context"
	"os"
	"strings"
	"testing"

	"account-service/service/internal/security"
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

func TestInitMigrationDefinesDefaultAdminSeed(t *testing.T) {
	sqlBytes, err := migrationFiles.ReadFile("migrations/000001_init.sql")
	if err != nil {
		t.Fatalf("read init migration: %v", err)
	}
	sql := string(sqlBytes)

	requiredFragments := []string{
		"insert into admin_users",
		"'admin'",
		"crypt('strongpass', gen_salt('bf'))",
		"'active'",
		"on conflict (username) do nothing",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("init migration missing default admin seed fragment %q", fragment)
		}
	}
}

func TestInitMigrationSeedsDefaultAdminUser(t *testing.T) {
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

	var username string
	var passwordHash string
	var status string
	err := pool.QueryRow(ctx, `
		select username, password_hash, status
		from admin_users
		where username = 'admin'
	`).Scan(&username, &passwordHash, &status)
	if err != nil {
		t.Fatalf("query default admin user: %v", err)
	}
	if username != "admin" {
		t.Fatalf("username = %q, want admin", username)
	}
	if passwordHash == "strongpass" {
		t.Fatal("password_hash must not store the plaintext password")
	}
	if !security.VerifyPassword("strongpass", passwordHash) {
		t.Fatal("expected strongpass to verify against seeded admin password hash")
	}
	if status != "active" {
		t.Fatalf("status = %q, want active", status)
	}
}
