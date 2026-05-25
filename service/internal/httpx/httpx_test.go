package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"account-service/service/internal/admin"
	"account-service/service/internal/callers"
	"account-service/service/internal/security"

	"github.com/gofiber/fiber/v3"
)

func TestJSONErrorIncludesRequestID(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/fail", func(c fiber.Ctx) error {
		return JSONError(c, http.StatusNotFound, "account_not_found", "Account not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	req.Header.Set("X-Request-ID", "request-id")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	if resp.Header.Get("X-Request-ID") != "request-id" {
		t.Fatalf("X-Request-ID = %q, want request-id", resp.Header.Get("X-Request-ID"))
	}
}

func TestAPIKeyAuthMiddleware(t *testing.T) {
	store := callers.NewMemoryStore()
	created, err := store.Create("worker", "test worker")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	app := fiber.New()
	app.Use(APIKeyAuth(store))
	app.Get("/protected", func(c fiber.Ctx) error {
		caller := CallerFromContext(c)
		return c.JSON(fiber.Map{"caller_id": caller.ID})
	})

	missingResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/protected", nil))
	if err != nil {
		t.Fatalf("missing app.Test() error = %v", err)
	}
	if missingResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("missing status = %d, want %d", missingResp.StatusCode, http.StatusUnauthorized)
	}

	wrongReq := httptest.NewRequest(http.MethodGet, "/protected", nil)
	wrongReq.Header.Set("Authorization", "Bearer wrong")
	wrongResp, err := app.Test(wrongReq)
	if err != nil {
		t.Fatalf("wrong app.Test() error = %v", err)
	}
	if wrongResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong status = %d, want %d", wrongResp.StatusCode, http.StatusUnauthorized)
	}

	validReq := httptest.NewRequest(http.MethodGet, "/protected", nil)
	validReq.Header.Set("Authorization", "Bearer "+created.PlaintextAPIKey)
	validResp, err := app.Test(validReq)
	if err != nil {
		t.Fatalf("valid app.Test() error = %v", err)
	}
	if validResp.StatusCode != http.StatusOK {
		t.Fatalf("valid status = %d, want %d", validResp.StatusCode, http.StatusOK)
	}

	store.Disable(created.Caller.ID)
	disabledReq := httptest.NewRequest(http.MethodGet, "/protected", nil)
	disabledReq.Header.Set("Authorization", "Bearer "+created.PlaintextAPIKey)
	disabledResp, err := app.Test(disabledReq)
	if err != nil {
		t.Fatalf("disabled app.Test() error = %v", err)
	}
	if disabledResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("disabled status = %d, want %d", disabledResp.StatusCode, http.StatusUnauthorized)
	}
}

func TestCORSMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(CORS([]string{"https://admin.example.com"}))
	app.Get("/resource", func(c fiber.Ctx) error { return c.SendStatus(http.StatusNoContent) })

	allowedReq := httptest.NewRequest(http.MethodOptions, "/resource", nil)
	allowedReq.Header.Set("Origin", "https://admin.example.com")
	allowedReq.Header.Set("Access-Control-Request-Method", "GET")
	allowedResp, err := app.Test(allowedReq)
	if err != nil {
		t.Fatalf("allowed app.Test() error = %v", err)
	}
	if allowedResp.Header.Get("Access-Control-Allow-Origin") != "https://admin.example.com" {
		t.Fatalf("Allow-Origin = %q, want allowed origin", allowedResp.Header.Get("Access-Control-Allow-Origin"))
	}
	if allowedResp.Header.Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("Allow-Credentials = %q, want true", allowedResp.Header.Get("Access-Control-Allow-Credentials"))
	}

	blockedReq := httptest.NewRequest(http.MethodGet, "/resource", nil)
	blockedReq.Header.Set("Origin", "https://evil.example.com")
	blockedResp, err := app.Test(blockedReq)
	if err != nil {
		t.Fatalf("blocked app.Test() error = %v", err)
	}
	if blockedResp.Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("blocked Allow-Origin = %q, want empty", blockedResp.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestAdminSessionMiddleware(t *testing.T) {
	passwordHash, err := security.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	store := admin.NewMemoryStore(time.Hour)
	store.AddUser(admin.User{ID: "admin-id", Username: "admin", PasswordHash: passwordHash, Status: "active"})
	service := admin.NewService(store, "session-secret", 15*time.Minute, 7*24*time.Hour)
	app := fiber.New()
	app.Post("/login", service.LoginHandler())
	app.Use(AdminSession(service))
	app.Get("/admin", func(c fiber.Ctx) error { return c.SendStatus(http.StatusNoContent) })

	loginResp, err := app.Test(jsonRequest(http.MethodPost, "/login", `{"username":"admin","password":"password123"}`))
	if err != nil {
		t.Fatalf("login app.Test() error = %v", err)
	}
	var loginBody admin.AuthResponse
	if err := json.NewDecoder(loginResp.Body).Decode(&loginBody); err != nil {
		t.Fatalf("decode login body: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+loginBody.AccessToken)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("admin app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestHashAPIKeyUsedByCallerStore(t *testing.T) {
	key := "acct_test"
	hash := security.HashAPIKey(key)
	if !security.VerifyAPIKey(key, hash) {
		t.Fatal("expected hash to verify")
	}
}

func jsonRequest(method string, path string, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
