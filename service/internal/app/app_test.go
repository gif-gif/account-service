package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"account-service/service/internal/accounts"
	"account-service/service/internal/audit"
	"account-service/service/internal/callers"
	"account-service/service/internal/leases"
	"account-service/service/internal/security"
)

func TestNewRegistersAccountRoutes(t *testing.T) {
	codec, err := security.NewCredentialCodec("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("NewCredentialCodec() error = %v", err)
	}
	accountService := accounts.NewService(accounts.NewMemoryRepository(codec), codec, audit.NewMemoryWriter())
	app := New(Options{AccountService: accountService})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/query", strings.NewReader(`{"limit":10}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestNewRegistersLeaseRoutes(t *testing.T) {
	codec, err := security.NewCredentialCodec("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("NewCredentialCodec() error = %v", err)
	}
	accountService := accounts.NewService(accounts.NewMemoryRepository(codec), codec, audit.NewMemoryWriter())
	leaseService := leases.NewService(accountService, 900000000000, 7200000000000, audit.NewMemoryWriter())
	app := New(Options{LeaseService: leaseService})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leases", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestNewRegistersCallerRoutesAndCORS(t *testing.T) {
	app := New(Options{
		CallerStore: callers.NewMemoryStore(),
		CORSOrigins: []string{"https://admin.example.com"},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-keys", strings.NewReader(`{"name":"worker","description":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://admin.example.com")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "https://admin.example.com" {
		t.Fatalf("Allow-Origin = %q, want frontend origin", resp.Header.Get("Access-Control-Allow-Origin"))
	}
	if resp.Header.Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID response header")
	}
}
