package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"account-service/service/internal/security"

	"github.com/gofiber/fiber/v3"
)

func TestLoginCurrentUserRefreshAndLogoutWithJWT(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app, newTestService(t, 15*time.Minute, 7*24*time.Hour))

	loginResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/admin/login", `{"username":"admin","password":"password123"}`))
	if err != nil {
		t.Fatalf("login app.Test() error = %v", err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d, want %d", loginResp.StatusCode, http.StatusOK)
	}
	if cookie := loginResp.Header.Get("Set-Cookie"); cookie != "" {
		t.Fatalf("Set-Cookie = %q, want empty because JWT credentials are returned in JSON", cookie)
	}
	loginBody := decodeAuthResponse(t, loginResp)
	if loginBody.AccessToken == "" || loginBody.RefreshToken == "" {
		t.Fatalf("tokens = %#v, want accessToken and refreshToken", loginBody)
	}
	if loginBody.User.Username != "admin" {
		t.Fatalf("username = %q, want admin", loginBody.User.Username)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginBody.AccessToken)
	meResp, err := app.Test(meReq)
	if err != nil {
		t.Fatalf("me app.Test() error = %v", err)
	}
	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("me status = %d, want %d", meResp.StatusCode, http.StatusOK)
	}

	refreshResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/admin/refresh", `{"refreshToken":"`+loginBody.RefreshToken+`"}`))
	if err != nil {
		t.Fatalf("refresh app.Test() error = %v", err)
	}
	if refreshResp.StatusCode != http.StatusOK {
		t.Fatalf("refresh status = %d, want %d", refreshResp.StatusCode, http.StatusOK)
	}
	refreshBody := decodeAuthResponse(t, refreshResp)
	if refreshBody.AccessToken == "" || refreshBody.RefreshToken == "" {
		t.Fatalf("refresh tokens = %#v, want accessToken and refreshToken", refreshBody)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+refreshBody.AccessToken)
	logoutResp, err := app.Test(logoutReq)
	if err != nil {
		t.Fatalf("logout app.Test() error = %v", err)
	}
	if logoutResp.StatusCode != http.StatusOK {
		t.Fatalf("logout status = %d, want %d", logoutResp.StatusCode, http.StatusOK)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app, newTestService(t, 15*time.Minute, 7*24*time.Hour))

	resp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/admin/login", `{"username":"admin","password":"wrong"}`))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestExpiredAccessTokenIsRejected(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app, newTestService(t, -time.Minute, 7*24*time.Hour))

	loginResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/admin/login", `{"username":"admin","password":"password123"}`))
	if err != nil {
		t.Fatalf("login app.Test() error = %v", err)
	}
	loginBody := decodeAuthResponse(t, loginResp)

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginBody.AccessToken)
	meResp, err := app.Test(meReq)
	if err != nil {
		t.Fatalf("me app.Test() error = %v", err)
	}
	if meResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", meResp.StatusCode, http.StatusUnauthorized)
	}
}

func TestExpiredRefreshTokenIsRejected(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app, newTestService(t, 15*time.Minute, -time.Minute))

	loginResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/admin/login", `{"username":"admin","password":"password123"}`))
	if err != nil {
		t.Fatalf("login app.Test() error = %v", err)
	}
	loginBody := decodeAuthResponse(t, loginResp)

	refreshResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/admin/refresh", `{"refreshToken":"`+loginBody.RefreshToken+`"}`))
	if err != nil {
		t.Fatalf("refresh app.Test() error = %v", err)
	}
	if refreshResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", refreshResp.StatusCode, http.StatusUnauthorized)
	}
}

func newTestService(t *testing.T, accessTTL time.Duration, refreshTTL time.Duration) *Service {
	t.Helper()
	passwordHash, err := security.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	store := NewMemoryStore(time.Hour)
	store.AddUser(User{ID: "admin-id", Username: "admin", PasswordHash: passwordHash, Status: "active"})
	return NewService(store, "session-secret", accessTTL, refreshTTL)
}

func decodeAuthResponse(t *testing.T, resp *http.Response) AuthResponse {
	t.Helper()
	var body AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

func jsonRequest(method string, path string, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
