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
	"account-service/service/internal/modelconfig"
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

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/missing-account", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+loginBody.AccessToken)
	deleteResp, err := app.Test(deleteReq)
	if err != nil {
		t.Fatalf("delete app.Test() error = %v", err)
	}
	if deleteResp.StatusCode != http.StatusNotFound {
		t.Fatalf("delete status = %d, want %d", deleteResp.StatusCode, http.StatusNotFound)
	}
	if contentType := deleteResp.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("delete Content-Type = %q, want JSON error", contentType)
	}
}

func TestNewExposesExternalRoutesWithAPIKey(t *testing.T) {
	codec, err := security.NewCredentialCodec("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("NewCredentialCodec() error = %v", err)
	}
	accountService := accounts.NewService(accounts.NewMemoryRepository(codec), codec, audit.NewMemoryWriter())
	createdAccount, err := accountService.Create(accounts.CreateAccountRequest{
		Username:            "worker@example.com",
		Password:            "plain-password",
		LoginURL:            "https://example.com/login",
		AccessToken:         "access-token",
		RefreshToken:        "refresh-token",
		Region:              "us",
		AccountType:         accounts.AccountTypeCodex,
		Status:              accounts.StatusActive,
		QuotaRemaining:      900,
		MaxConcurrentLeases: 1,
		Tags:                []string{"openai"},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	leaseService := leases.NewService(accountService, 15*time.Minute, 2*time.Hour, audit.NewMemoryWriter())
	callerStore := callers.NewMemoryStore()
	callerResult, err := callerStore.Create("worker", "external worker")
	if err != nil {
		t.Fatalf("Create caller error = %v", err)
	}
	adminService := newTestAdminService(t)
	modelConfigService := modelconfig.NewService(modelconfig.NewMemoryRepository([]modelconfig.Item{
		{Kind: modelconfig.KindFallbackModel, Key: "auto", DisplayOrder: 10},
		{Kind: modelconfig.KindModelAlias, Key: "claude-opus-4-7", Value: "claude-opus-4.7", DisplayOrder: 20},
	}))
	app := New(Options{
		AdminService:   adminService,
		AccountService: accountService,
		LeaseService:   leaseService,
		CallerStore:    callerStore,
		ModelConfig:    modelConfigService,
	})

	missingReq := httptest.NewRequest(http.MethodPost, "/api/v1/external/accounts/query", strings.NewReader(`{"limit":10}`))
	missingReq.Header.Set("Content-Type", "application/json")
	missingResp, err := app.Test(missingReq)
	if err != nil {
		t.Fatalf("missing api key app.Test() error = %v", err)
	}
	if missingResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("missing api key status = %d, want %d", missingResp.StatusCode, http.StatusUnauthorized)
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
	adminReq := httptest.NewRequest(http.MethodPost, "/api/v1/external/accounts/query", strings.NewReader(`{"limit":10}`))
	adminReq.Header.Set("Content-Type", "application/json")
	adminReq.Header.Set("Authorization", "Bearer "+loginBody.AccessToken)
	adminResp, err := app.Test(adminReq)
	if err != nil {
		t.Fatalf("admin token app.Test() error = %v", err)
	}
	if adminResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("admin token status = %d, want %d", adminResp.StatusCode, http.StatusUnauthorized)
	}

	queryReq := httptest.NewRequest(http.MethodPost, "/api/v1/external/accounts/query", strings.NewReader(`{"region":"us","account_type":"codex","statuses":["active"],"limit":10}`))
	queryReq.Header.Set("Content-Type", "application/json")
	queryReq.Header.Set("Authorization", "Bearer "+callerResult.PlaintextAPIKey)
	queryResp, err := app.Test(queryReq)
	if err != nil {
		t.Fatalf("query app.Test() error = %v", err)
	}
	if queryResp.StatusCode != http.StatusOK {
		t.Fatalf("query status = %d, want %d", queryResp.StatusCode, http.StatusOK)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/external/accounts?region=us&account_type=codex&status=active&limit=10", nil)
	listReq.Header.Set("Authorization", "Bearer "+callerResult.PlaintextAPIKey)
	listResp, err := app.Test(listReq)
	if err != nil {
		t.Fatalf("list app.Test() error = %v", err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listResp.StatusCode, http.StatusOK)
	}
	var listBody struct {
		Accounts []accounts.Account `json:"accounts"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listBody.Accounts) != 1 {
		t.Fatalf("list accounts length = %d, want 1", len(listBody.Accounts))
	}
	if listBody.Accounts[0].ID != createdAccount.ID {
		t.Fatalf("list account id = %q, want %q", listBody.Accounts[0].ID, createdAccount.ID)
	}

	acquireReq := httptest.NewRequest(http.MethodPost, "/api/v1/external/accounts/acquire", strings.NewReader(`{"region":"us","account_type":"codex","ttl_seconds":900,"caller_id":"spoofed-caller"}`))
	acquireReq.Header.Set("Content-Type", "application/json")
	acquireReq.Header.Set("Authorization", "Bearer "+callerResult.PlaintextAPIKey)
	acquireResp, err := app.Test(acquireReq)
	if err != nil {
		t.Fatalf("acquire app.Test() error = %v", err)
	}
	if acquireResp.StatusCode != http.StatusOK {
		t.Fatalf("acquire status = %d, want %d", acquireResp.StatusCode, http.StatusOK)
	}
	var leaseBody leases.Lease
	if err := json.NewDecoder(acquireResp.Body).Decode(&leaseBody); err != nil {
		t.Fatalf("decode acquire response: %v", err)
	}
	if leaseBody.CallerID != callerResult.Caller.ID {
		t.Fatalf("lease caller_id = %q, want authenticated caller %q", leaseBody.CallerID, callerResult.Caller.ID)
	}

	releaseReq := httptest.NewRequest(http.MethodPost, "/api/v1/external/accounts/release", strings.NewReader(`{"lease_id":"`+leaseBody.ID+`"}`))
	releaseReq.Header.Set("Content-Type", "application/json")
	releaseReq.Header.Set("Authorization", "Bearer "+callerResult.PlaintextAPIKey)
	releaseResp, err := app.Test(releaseReq)
	if err != nil {
		t.Fatalf("release app.Test() error = %v", err)
	}
	if releaseResp.StatusCode != http.StatusOK {
		t.Fatalf("release status = %d, want %d", releaseResp.StatusCode, http.StatusOK)
	}

	statusReq := httptest.NewRequest(http.MethodPost, "/api/v1/external/accounts/"+createdAccount.ID+"/status", strings.NewReader(`{"status":"disabled","reason":"maintenance"}`))
	statusReq.Header.Set("Content-Type", "application/json")
	statusReq.Header.Set("Authorization", "Bearer "+callerResult.PlaintextAPIKey)
	statusResp, err := app.Test(statusReq)
	if err != nil {
		t.Fatalf("status app.Test() error = %v", err)
	}
	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("status update status = %d, want %d", statusResp.StatusCode, http.StatusOK)
	}
	var statusBody struct {
		Account accounts.Account `json:"account"`
	}
	if err := json.NewDecoder(statusResp.Body).Decode(&statusBody); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if statusBody.Account.Status != accounts.StatusDisabled {
		t.Fatalf("updated status = %q, want %q", statusBody.Account.Status, accounts.StatusDisabled)
	}

	modelConfigReq := httptest.NewRequest(http.MethodGet, "/api/v1/external/model-config", nil)
	modelConfigReq.Header.Set("Authorization", "Bearer "+callerResult.PlaintextAPIKey)
	modelConfigResp, err := app.Test(modelConfigReq)
	if err != nil {
		t.Fatalf("model config app.Test() error = %v", err)
	}
	if modelConfigResp.StatusCode != http.StatusOK {
		t.Fatalf("model config status = %d, want %d", modelConfigResp.StatusCode, http.StatusOK)
	}
	var modelConfigBody struct {
		FallbackModels []struct {
			ModelID string `json:"model_id"`
		} `json:"fallback_models"`
		HiddenModels   map[string]string `json:"hidden_models"`
		ModelAliases   map[string]string `json:"model_aliases"`
		HiddenFromList []string          `json:"hidden_from_list"`
	}
	if err := json.NewDecoder(modelConfigResp.Body).Decode(&modelConfigBody); err != nil {
		t.Fatalf("decode model config response: %v", err)
	}
	if modelConfigBody.FallbackModels[0].ModelID != "auto" {
		t.Fatalf("model config fallback first model = %q, want auto", modelConfigBody.FallbackModels[0].ModelID)
	}
	if modelConfigBody.ModelAliases["claude-opus-4-7"] != "claude-opus-4.7" {
		t.Fatalf("model config alias claude-opus-4-7 = %q", modelConfigBody.ModelAliases["claude-opus-4-7"])
	}
}

func TestNewCanDisableExternalAPIKeyAuth(t *testing.T) {
	authEnabled := false
	app := New(Options{
		CallerStore:               callers.NewMemoryStore(),
		ExternalAPIKeyAuthEnabled: &authEnabled,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/external/model-config", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
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
