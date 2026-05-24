package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"account-service/service/internal/accounts"
	"account-service/service/internal/audit"
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
