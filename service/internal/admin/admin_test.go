package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"account-service/service/internal/security"

	"github.com/gofiber/fiber/v3"
)

func TestLoginCurrentUserAndLogout(t *testing.T) {
	passwordHash, err := security.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	store := NewMemoryStore(time.Hour)
	store.AddUser(User{ID: "admin-id", Username: "admin", PasswordHash: passwordHash, Status: "active"})

	app := fiber.New()
	RegisterRoutes(app, NewService(store, "session-secret", true))

	loginResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/admin/login", `{"username":"admin","password":"password123"}`))
	if err != nil {
		t.Fatalf("login app.Test() error = %v", err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d, want %d", loginResp.StatusCode, http.StatusOK)
	}
	cookie := loginResp.Header.Get("Set-Cookie")
	if !strings.Contains(cookie, "account_admin_session=") {
		t.Fatalf("Set-Cookie = %q, want admin session cookie", cookie)
	}
	lowerCookie := strings.ToLower(cookie)
	if !strings.Contains(lowerCookie, "httponly") || !strings.Contains(lowerCookie, "secure") || !strings.Contains(cookie, "SameSite=Lax") {
		t.Fatalf("Set-Cookie = %q, want HttpOnly, Secure, SameSite=Lax", cookie)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	meReq.Header.Set("Cookie", cookie)
	meResp, err := app.Test(meReq)
	if err != nil {
		t.Fatalf("me app.Test() error = %v", err)
	}
	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("me status = %d, want %d", meResp.StatusCode, http.StatusOK)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/logout", nil)
	logoutReq.Header.Set("Cookie", cookie)
	logoutResp, err := app.Test(logoutReq)
	if err != nil {
		t.Fatalf("logout app.Test() error = %v", err)
	}
	if logoutResp.StatusCode != http.StatusOK {
		t.Fatalf("logout status = %d, want %d", logoutResp.StatusCode, http.StatusOK)
	}

	meAfterLogoutReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	meAfterLogoutReq.Header.Set("Cookie", cookie)
	meAfterLogoutResp, err := app.Test(meAfterLogoutReq)
	if err != nil {
		t.Fatalf("me after logout app.Test() error = %v", err)
	}
	if meAfterLogoutResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("me after logout status = %d, want %d", meAfterLogoutResp.StatusCode, http.StatusUnauthorized)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	passwordHash, err := security.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	store := NewMemoryStore(time.Hour)
	store.AddUser(User{ID: "admin-id", Username: "admin", PasswordHash: passwordHash, Status: "active"})

	app := fiber.New()
	RegisterRoutes(app, NewService(store, "session-secret", true))

	resp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/admin/login", `{"username":"admin","password":"wrong"}`))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestExpiredSessionIsRejected(t *testing.T) {
	passwordHash, err := security.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	store := NewMemoryStore(-time.Minute)
	store.AddUser(User{ID: "admin-id", Username: "admin", PasswordHash: passwordHash, Status: "active"})

	app := fiber.New()
	RegisterRoutes(app, NewService(store, "session-secret", true))

	loginResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/admin/login", `{"username":"admin","password":"password123"}`))
	if err != nil {
		t.Fatalf("login app.Test() error = %v", err)
	}
	cookie := loginResp.Header.Get("Set-Cookie")

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	meReq.Header.Set("Cookie", cookie)
	meResp, err := app.Test(meReq)
	if err != nil {
		t.Fatalf("me app.Test() error = %v", err)
	}
	if meResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", meResp.StatusCode, http.StatusUnauthorized)
	}
}

func jsonRequest(method string, path string, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
