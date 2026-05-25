package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"account-service/service/internal/config"
)

func TestBuildAppRegistersDefaultAdminLogin(t *testing.T) {
	cfg := config.Config{
		SecretEncryptionKey: "0123456789abcdef0123456789abcdef",
		AdminSessionSecret:  "session-secret",
		DefaultLeaseTTL:     900000000000,
		MaxLeaseTTL:         7200000000000,
		JWTAccessTokenTTL:   900000000000,
		JWTRefreshTokenTTL:  604800000000000,
	}
	app, err := buildApp(cfg)
	if err != nil {
		t.Fatalf("buildApp() error = %v", err)
	}

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
