package accounts

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"account-service/service/internal/audit"
	"account-service/service/internal/security"

	"github.com/gofiber/fiber/v3"
)

func TestRepositoryCreateGetUpdateAndQuery(t *testing.T) {
	codec := mustCodec(t)
	repo := NewMemoryRepository(codec)
	svc := NewService(repo, codec, audit.NewMemoryWriter())

	created, err := svc.Create(CreateAccountRequest{
		Username:            "user@example.com",
		Password:            "plain-password",
		LoginURL:            "https://example.com/login",
		AccessToken:         "access-token",
		RefreshToken:        "refresh-token",
		Region:              "us",
		AccountType:         AccountTypeCodex,
		Status:              StatusActive,
		QuotaTotal:          1000,
		QuotaUsed:           100,
		QuotaRemaining:      900,
		MaxConcurrentLeases: 2,
		Tags:                []string{"openai", "paid"},
		Notes:               "first account",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Password != "plain-password" || created.AccessToken != "access-token" || created.RefreshToken != "refresh-token" {
		t.Fatalf("created credentials were not decrypted in response: %#v", created)
	}

	stored, ok := repo.Raw(created.ID)
	if !ok {
		t.Fatal("expected raw account to be stored")
	}
	if stored.PasswordEncrypted == "plain-password" || stored.AccessTokenEncrypted == "access-token" || stored.RefreshTokenEncrypted == "refresh-token" {
		t.Fatal("expected sensitive fields to be encrypted at rest")
	}

	updated, err := svc.Update(created.ID, UpdateAccountRequest{
		Status:         ptr(StatusTokenExpired),
		QuotaRemaining: ptrInt64(10),
		Tags:           []string{"openai"},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Status != StatusTokenExpired {
		t.Fatalf("Status = %q, want %q", updated.Status, StatusTokenExpired)
	}
	if updated.QuotaRemaining != 10 {
		t.Fatalf("QuotaRemaining = %d, want 10", updated.QuotaRemaining)
	}

	results, err := svc.Query(QueryRequest{
		Region:            "us",
		AccountType:       AccountTypeCodex,
		Statuses:          []Status{StatusTokenExpired},
		Tags:              []string{"openai"},
		MinQuotaRemaining: 1,
		Limit:             10,
	})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(results) != 1 || results[0].ID != created.ID {
		t.Fatalf("Query() = %#v, want created account", results)
	}
}

func TestRejectsInvalidStatus(t *testing.T) {
	codec := mustCodec(t)
	svc := NewService(NewMemoryRepository(codec), codec, audit.NewMemoryWriter())

	_, err := svc.Create(CreateAccountRequest{
		Username:            "user@example.com",
		Password:            "plain-password",
		LoginURL:            "https://example.com/login",
		AccessToken:         "access-token",
		RefreshToken:        "refresh-token",
		Region:              "us",
		AccountType:         AccountTypeCodex,
		Status:              "bad-status",
		MaxConcurrentLeases: 1,
	})
	if err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestRejectsInvalidAccountType(t *testing.T) {
	codec := mustCodec(t)
	svc := NewService(NewMemoryRepository(codec), codec, audit.NewMemoryWriter())

	_, err := svc.Create(CreateAccountRequest{
		Username:            "user@example.com",
		Password:            "plain-password",
		LoginURL:            "https://example.com/login",
		AccessToken:         "access-token",
		RefreshToken:        "refresh-token",
		Region:              "us",
		AccountType:         "pro",
		Status:              StatusActive,
		MaxConcurrentLeases: 1,
	})
	if err == nil {
		t.Fatal("expected invalid account type error")
	}
}

func TestHandlersExposeAccountAPI(t *testing.T) {
	codec := mustCodec(t)
	repo := NewMemoryRepository(codec)
	svc := NewService(repo, codec, audit.NewMemoryWriter())
	app := fiber.New()
	RegisterRoutes(app, svc)

	createResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/accounts", `{
		"username":"user@example.com",
		"password":"plain-password",
		"login_url":"https://example.com/login",
		"access_token":"access-token",
		"refresh_token":"refresh-token",
		"region":"us",
		"account_type":"codex",
		"status":"active",
		"quota_total":1000,
		"quota_used":100,
		"quota_remaining":900,
		"max_concurrent_leases":2,
		"tags":["openai"]
	}`))
	if err != nil {
		t.Fatalf("create app.Test() error = %v", err)
	}
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResp.StatusCode, http.StatusCreated)
	}
	accountID := createResp.Header.Get("X-Test-Account-ID")
	if accountID == "" {
		t.Fatal("expected X-Test-Account-ID header")
	}

	getResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+accountID, nil))
	if err != nil {
		t.Fatalf("get app.Test() error = %v", err)
	}
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getResp.StatusCode, http.StatusOK)
	}

	queryResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/accounts/query", `{"region":"us","account_type":"codex","statuses":["active"],"tags":["openai"],"min_quota_remaining":1,"limit":10}`))
	if err != nil {
		t.Fatalf("query app.Test() error = %v", err)
	}
	if queryResp.StatusCode != http.StatusOK {
		t.Fatalf("query status = %d, want %d", queryResp.StatusCode, http.StatusOK)
	}

	statusResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/accounts/"+accountID+"/status", `{"status":"token_expired","reason":"refresh failed"}`))
	if err != nil {
		t.Fatalf("status app.Test() error = %v", err)
	}
	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("status update status = %d, want %d", statusResp.StatusCode, http.StatusOK)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/"+accountID, nil)
	deleteResp, err := app.Test(deleteReq)
	if err != nil {
		t.Fatalf("delete app.Test() error = %v", err)
	}
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("delete status = %d, want %d", deleteResp.StatusCode, http.StatusOK)
	}

	getDeletedResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+accountID, nil))
	if err != nil {
		t.Fatalf("get deleted app.Test() error = %v", err)
	}
	if getDeletedResp.StatusCode != http.StatusNotFound {
		t.Fatalf("get deleted status = %d, want %d", getDeletedResp.StatusCode, http.StatusNotFound)
	}
}

func mustCodec(t *testing.T) security.CredentialCodec {
	t.Helper()
	codec, err := security.NewCredentialCodec("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("NewCredentialCodec() error = %v", err)
	}
	return codec
}

func ptr(status Status) *Status {
	return &status
}

func ptrInt64(value int64) *int64 {
	return &value
}

func jsonRequest(method string, path string, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
