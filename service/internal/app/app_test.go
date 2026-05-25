package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"account-service/service/internal/accounts"
	"account-service/service/internal/admin"
	"account-service/service/internal/audit"
	"account-service/service/internal/callers"
	"account-service/service/internal/leases"
	"account-service/service/internal/security"
)

func TestNewRegistersAdminRoutes(t *testing.T) {
	adminService := newTestAdminService(t)
	app := New(Options{AdminService: adminService})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", strings.NewReader(`{"username":"admin","password":"strongpass"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestNewProtectsManagedRoutesWithAdminAccessToken(t *testing.T) {
	codec, err := security.NewCredentialCodec("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("NewCredentialCodec() error = %v", err)
	}
	accountService := accounts.NewService(accounts.NewMemoryRepository(codec), codec, audit.NewMemoryWriter())
	app := New(Options{AdminService: newTestAdminService(t), AccountService: accountService})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/query", strings.NewReader(`{"limit":10}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unauthorized app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", strings.NewReader(`{"username":"admin","password":"strongpass"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq)
	if err != nil {
		t.Fatalf("login app.Test() error = %v", err)
	}
	var loginBody admin.AuthResponse
	if err := json.NewDecoder(loginResp.Body).Decode(&loginBody); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginBody.AccessToken == "" {
		t.Fatal("expected login response to include accessToken")
	}

	authorizedReq := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/query", strings.NewReader(`{"limit":10}`))
	authorizedReq.Header.Set("Content-Type", "application/json")
	authorizedReq.Header.Set("Authorization", "Bearer "+loginBody.AccessToken)
	authorizedResp, err := app.Test(authorizedReq)
	if err != nil {
		t.Fatalf("authorized app.Test() error = %v", err)
	}
	if authorizedResp.StatusCode != http.StatusOK {
		t.Fatalf("authorized status = %d, want %d", authorizedResp.StatusCode, http.StatusOK)
	}
}

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

func newTestAdminService(t *testing.T) *admin.Service {
	t.Helper()
	passwordHash, err := security.HashPassword("strongpass")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	store := admin.NewMemoryStore(time.Hour)
	store.AddUser(admin.User{ID: "admin-id", Username: "admin", PasswordHash: passwordHash, Status: "active"})
	return admin.NewService(store, "session-secret", 15*time.Minute, 7*24*time.Hour)
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
