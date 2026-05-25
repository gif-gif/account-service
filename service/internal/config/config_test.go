package config

import (
	"testing"
	"time"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SECRET_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("ADMIN_SESSION_SECRET", "admin-session-secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://account:account@localhost:5432/account?sslmode=disable")
	t.Setenv("SECRET_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("ADMIN_SESSION_SECRET", "admin-session-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPHost != "127.0.0.1" {
		t.Fatalf("HTTPHost = %q, want %q", cfg.HTTPHost, "127.0.0.1")
	}
	if cfg.HTTPPort != 8000 {
		t.Fatalf("HTTPPort = %d, want %d", cfg.HTTPPort, 8000)
	}
	if cfg.DefaultLeaseTTL != 15*time.Minute {
		t.Fatalf("DefaultLeaseTTL = %s, want 15m", cfg.DefaultLeaseTTL)
	}
	if cfg.MaxLeaseTTL != 2*time.Hour {
		t.Fatalf("MaxLeaseTTL = %s, want 2h", cfg.MaxLeaseTTL)
	}
	if cfg.LeaseCleanupInterval != time.Minute {
		t.Fatalf("LeaseCleanupInterval = %s, want 1m", cfg.LeaseCleanupInterval)
	}
	if cfg.JWTAccessTokenTTL != 15*time.Minute {
		t.Fatalf("JWTAccessTokenTTL = %s, want 15m", cfg.JWTAccessTokenTTL)
	}
	if cfg.JWTRefreshTokenTTL != 7*24*time.Hour {
		t.Fatalf("JWTRefreshTokenTTL = %s, want 168h", cfg.JWTRefreshTokenTTL)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.HealthCheckDatabaseTimeout != 3*time.Second {
		t.Fatalf("HealthCheckDatabaseTimeout = %s, want 3s", cfg.HealthCheckDatabaseTimeout)
	}
}

func TestLoadParsesCORSAllowedOrigins(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://account:account@localhost:5432/account?sslmode=disable")
	t.Setenv("SECRET_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("ADMIN_SESSION_SECRET", "admin-session-secret")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://admin.example.com, http://localhost:5173 ,,")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := []string{"https://admin.example.com", "http://localhost:5173"}
	if len(cfg.CORSAllowedOrigins) != len(want) {
		t.Fatalf("CORSAllowedOrigins = %#v, want %#v", cfg.CORSAllowedOrigins, want)
	}
	for i := range want {
		if cfg.CORSAllowedOrigins[i] != want[i] {
			t.Fatalf("CORSAllowedOrigins[%d] = %q, want %q", i, cfg.CORSAllowedOrigins[i], want[i])
		}
	}
}

func TestLoadRejectsInvalidTTL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://account:account@localhost:5432/account?sslmode=disable")
	t.Setenv("SECRET_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("ADMIN_SESSION_SECRET", "admin-session-secret")
	t.Setenv("DEFAULT_LEASE_TTL_SECONDS", "7201")
	t.Setenv("MAX_LEASE_TTL_SECONDS", "7200")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when default lease TTL is greater than max lease TTL")
	}
}
