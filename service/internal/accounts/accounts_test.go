package accounts

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"account-service/service/internal/audit"
	"account-service/service/internal/logging"
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

func TestCreateDefaultsStatusToDisabled(t *testing.T) {
	codec := mustCodec(t)
	svc := NewService(NewMemoryRepository(codec), codec, audit.NewMemoryWriter())

	created, err := svc.Create(CreateAccountRequest{
		Username:            "user@example.com",
		Password:            "plain-password",
		LoginURL:            "https://example.com/login",
		AccessToken:         "access-token",
		RefreshToken:        "refresh-token",
		Region:              "us",
		AccountType:         AccountTypeKiroAWS,
		MaxConcurrentLeases: 1,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Status != StatusDisabled {
		t.Fatalf("Status = %q, want %q", created.Status, StatusDisabled)
	}
}

func TestAccountTypeKiroVariants(t *testing.T) {
	codec := mustCodec(t)
	svc := NewService(NewMemoryRepository(codec), codec, audit.NewMemoryWriter())

	for _, accountType := range []AccountType{AccountTypeKiroAWS, AccountTypeKiroOffical} {
		if _, err := svc.Create(CreateAccountRequest{
			Username:            "user@example.com",
			Password:            "plain-password",
			LoginURL:            "https://example.com/login",
			AccessToken:         "access-token",
			RefreshToken:        "refresh-token",
			Region:              "us",
			AccountType:         accountType,
			MaxConcurrentLeases: 1,
		}); err != nil {
			t.Fatalf("Create() with account type %q error = %v", accountType, err)
		}
	}

	if _, err := svc.Create(CreateAccountRequest{
		Username:            "user@example.com",
		Password:            "plain-password",
		LoginURL:            "https://example.com/login",
		AccessToken:         "access-token",
		RefreshToken:        "refresh-token",
		Region:              "us",
		AccountType:         "kiro",
		MaxConcurrentLeases: 1,
	}); err == nil {
		t.Fatal("expected legacy kiro account type to be rejected")
	}
}

func TestKiroLoginUpdatesAccountOnSuccess(t *testing.T) {
	codec := mustCodec(t)
	svc := NewService(NewMemoryRepository(codec), codec, audit.NewMemoryWriter())
	runner := &fakeKiroLoginRunner{
		success: true,
		config: &KiroCliConfig{
			AccessToken:  "kiro-access",
			RefreshToken: "kiro-refresh",
			ExpiresAt:    time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC),
			ProfileARN:   "arn:aws:iam::123456789012:role/Kiro",
			AuthMethod:   "device",
			Provider:     "google",
		},
		done: make(chan struct{}),
	}
	svc.SetKiroLoginRunner(runner)
	account := createTestAccount(t, svc, StatusDisabled)

	result, err := svc.StartKiroLogin(account.ID)
	if err != nil {
		t.Fatalf("StartKiroLogin() error = %v", err)
	}
	if result.Status != "running" {
		t.Fatalf("login status = %q, want running", result.Status)
	}
	waitForFakeRunner(t, runner)
	waitForAccountStatus(t, svc, account.ID, StatusActive)

	updated, err := svc.Get(account.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if updated.Status != StatusActive {
		t.Fatalf("Status = %q, want %q", updated.Status, StatusActive)
	}
	if updated.AccessToken != "kiro-access" || updated.RefreshToken != "kiro-refresh" {
		t.Fatalf("tokens were not updated: access=%q refresh=%q", updated.AccessToken, updated.RefreshToken)
	}
	if updated.KiroProfileARN != "arn:aws:iam::123456789012:role/Kiro" {
		t.Fatalf("KiroProfileARN = %q", updated.KiroProfileARN)
	}
	if updated.KiroAuthMethod != "device" || updated.KiroProvider != "google" {
		t.Fatalf("kiro config was not updated: %#v", updated)
	}
	if updated.KiroExpiresAt == nil || !updated.KiroExpiresAt.Equal(runner.config.ExpiresAt) {
		t.Fatalf("KiroExpiresAt = %v, want %v", updated.KiroExpiresAt, runner.config.ExpiresAt)
	}
}

func TestKiroLoginMarksAccountFailedOnFailure(t *testing.T) {
	codec := mustCodec(t)
	svc := NewService(NewMemoryRepository(codec), codec, audit.NewMemoryWriter())
	runner := &fakeKiroLoginRunner{success: false, done: make(chan struct{})}
	svc.SetKiroLoginRunner(runner)
	account := createTestAccount(t, svc, StatusDisabled)

	if _, err := svc.StartKiroLogin(account.ID); err != nil {
		t.Fatalf("StartKiroLogin() error = %v", err)
	}
	waitForFakeRunner(t, runner)
	waitForAccountStatus(t, svc, account.ID, StatusLoginFailed)

	updated, err := svc.Get(account.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if updated.Status != StatusLoginFailed {
		t.Fatalf("Status = %q, want %q", updated.Status, StatusLoginFailed)
	}
}

func TestKiroLoginLogsUpdateErrors(t *testing.T) {
	var logs bytes.Buffer
	logging.SetDefault(logging.New(&logs, "debug"))
	t.Cleanup(func() {
		logging.SetDefault(logging.New(&bytes.Buffer{}, "info"))
	})

	codec := mustCodec(t)
	svc := NewService(NewMemoryRepository(codec), codec, audit.NewMemoryWriter())
	release := make(chan struct{})
	runner := &fakeKiroLoginRunner{
		success: true,
		config: &KiroCliConfig{
			AccessToken:  "kiro-access",
			RefreshToken: "kiro-refresh",
			ExpiresAt:    time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC),
			ProfileARN:   "arn:aws:iam::123456789012:role/Kiro",
			AuthMethod:   "device",
			Provider:     "google",
		},
		done: make(chan struct{}),
		wait: release,
	}
	svc.SetKiroLoginRunner(runner)
	account := createTestAccount(t, svc, StatusDisabled)

	if _, err := svc.StartKiroLogin(account.ID); err != nil {
		t.Fatalf("StartKiroLogin() error = %v", err)
	}
	if err := svc.Delete(account.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	close(release)
	waitForFakeRunner(t, runner)

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(logs.String(), "update kiro login account") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("logs = %q, want update error log", logs.String())
}

func TestCancelKiroLoginCallsRunnerCancel(t *testing.T) {
	codec := mustCodec(t)
	svc := NewService(NewMemoryRepository(codec), codec, audit.NewMemoryWriter())
	runner := &fakeKiroLoginRunner{success: false, done: make(chan struct{})}
	svc.SetKiroLoginRunner(runner)

	if err := svc.CancelKiroLogin("account-id"); err != nil {
		t.Fatalf("CancelKiroLogin() error = %v", err)
	}
	if runner.cancelCalls != 1 {
		t.Fatalf("cancelCalls = %d, want 1", runner.cancelCalls)
	}
}

func TestKiroLoginRunningUsesRunnerState(t *testing.T) {
	codec := mustCodec(t)
	svc := NewService(NewMemoryRepository(codec), codec, audit.NewMemoryWriter())
	runner := &fakeKiroLoginRunner{running: true}
	svc.SetKiroLoginRunner(runner)

	running := svc.KiroLoginRunning("account-id")
	if !running {
		t.Fatal("KiroLoginRunning() = false, want true")
	}

	runner.running = false
	running = svc.KiroLoginRunning("account-id")
	if running {
		t.Fatal("KiroLoginRunning() = true, want false")
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

func TestHandlersExposeKiroLoginAPI(t *testing.T) {
	codec := mustCodec(t)
	svc := NewService(NewMemoryRepository(codec), codec, audit.NewMemoryWriter())
	runner := &fakeKiroLoginRunner{success: false, done: make(chan struct{})}
	svc.SetKiroLoginRunner(runner)
	account := createTestAccount(t, svc, StatusDisabled)
	app := fiber.New()
	RegisterRoutes(app, svc)

	loginResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/accounts/"+account.ID+"/kiroLogin", `{}`))
	if err != nil {
		t.Fatalf("kiroLogin app.Test() error = %v", err)
	}
	if loginResp.StatusCode != http.StatusAccepted {
		t.Fatalf("kiroLogin status = %d, want %d", loginResp.StatusCode, http.StatusAccepted)
	}
	var loginBody KiroLoginResult
	if err := json.NewDecoder(loginResp.Body).Decode(&loginBody); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginBody.AccountID != account.ID || loginBody.Status != "running" {
		t.Fatalf("login response = %#v", loginBody)
	}
	waitForFakeRunner(t, runner)
	if len(runner.awsLoginAccounts) != 1 {
		t.Fatalf("awsLoginAccounts len = %d, want 1", len(runner.awsLoginAccounts))
	}
	if runner.awsLoginAccounts[0].ID != account.ID {
		t.Fatalf("aws login account ID = %q, want %q", runner.awsLoginAccounts[0].ID, account.ID)
	}
	if runner.awsLoginAccounts[0].AccountType != AccountTypeKiroAWS {
		t.Fatalf("aws login account type = %q, want %q", runner.awsLoginAccounts[0].AccountType, AccountTypeKiroAWS)
	}

	cancelResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/accounts/"+account.ID+"/cancelKiroLogin", `{}`))
	if err != nil {
		t.Fatalf("cancelKiroLogin app.Test() error = %v", err)
	}
	if cancelResp.StatusCode != http.StatusOK {
		t.Fatalf("cancel status = %d, want %d", cancelResp.StatusCode, http.StatusOK)
	}
	if runner.cancelCalls != 1 {
		t.Fatalf("cancelCalls = %d, want 1", runner.cancelCalls)
	}

	runner.running = true
	runningResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+account.ID+"/kiroLogin/running", nil))
	if err != nil {
		t.Fatalf("kiroLogin running app.Test() error = %v", err)
	}
	if runningResp.StatusCode != http.StatusOK {
		t.Fatalf("running status = %d, want %d", runningResp.StatusCode, http.StatusOK)
	}
	var runningBody struct {
		Running bool `json:"running"`
	}
	if err := json.NewDecoder(runningResp.Body).Decode(&runningBody); err != nil {
		t.Fatalf("decode running response: %v", err)
	}
	if !runningBody.Running {
		t.Fatalf("running response = %#v, want running true", runningBody)
	}
}

func TestStartKiroLoginUsesOfficialRunnerForOfficialAccount(t *testing.T) {
	codec := mustCodec(t)
	svc := NewService(NewMemoryRepository(codec), codec, audit.NewMemoryWriter())
	runner := &fakeKiroLoginRunner{success: false, done: make(chan struct{})}
	svc.SetKiroLoginRunner(runner)
	account, err := svc.Create(CreateAccountRequest{
		Username:            "official@example.com",
		Password:            "plain-password",
		LoginURL:            "https://app.kiro.dev/account/device",
		Region:              "us-east-1",
		AccountType:         AccountTypeKiroOffical,
		Status:              StatusDisabled,
		MaxConcurrentLeases: 1,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := svc.StartKiroLogin(account.ID); err != nil {
		t.Fatalf("StartKiroLogin() error = %v", err)
	}
	waitForFakeRunner(t, runner)

	if len(runner.officialLoginAccounts) != 1 {
		t.Fatalf("officialLoginAccounts len = %d, want 1", len(runner.officialLoginAccounts))
	}
	if runner.officialLoginAccounts[0].ID != account.ID {
		t.Fatalf("official login account ID = %q, want %q", runner.officialLoginAccounts[0].ID, account.ID)
	}
	if len(runner.awsLoginAccounts) != 0 {
		t.Fatalf("awsLoginAccounts len = %d, want 0", len(runner.awsLoginAccounts))
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

type fakeKiroLoginRunner struct {
	success               bool
	config                *KiroCliConfig
	done                  chan struct{}
	wait                  chan struct{}
	cancelCalls           int
	running               bool
	officialLoginAccounts []Account
	awsLoginAccounts      []Account
}

func (runner *fakeKiroLoginRunner) KiroCliLogin(account Account) (bool, *KiroCliConfig) {
	runner.officialLoginAccounts = append(runner.officialLoginAccounts, account)
	if runner.wait != nil {
		<-runner.wait
	}
	defer close(runner.done)
	return runner.success, runner.config
}

func (runner *fakeKiroLoginRunner) KiroCliLoginByAws(account Account) (bool, *KiroCliConfig) {
	runner.awsLoginAccounts = append(runner.awsLoginAccounts, account)
	if runner.wait != nil {
		<-runner.wait
	}
	defer close(runner.done)
	return runner.success, runner.config
}

func (runner *fakeKiroLoginRunner) Cancel() {
	runner.cancelCalls++
}

func (runner *fakeKiroLoginRunner) Running() bool {
	return runner.running
}

func createTestAccount(t *testing.T, svc *Service, status Status) Account {
	t.Helper()
	account, err := svc.Create(CreateAccountRequest{
		Username:            "user@example.com",
		Password:            "plain-password",
		LoginURL:            "https://example.com/login",
		AccessToken:         "access-token",
		RefreshToken:        "refresh-token",
		Region:              "us",
		AccountType:         AccountTypeKiroAWS,
		Status:              status,
		MaxConcurrentLeases: 1,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	return account
}

func waitForFakeRunner(t *testing.T, runner *fakeKiroLoginRunner) {
	t.Helper()
	select {
	case <-runner.done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for fake kiro runner")
	}
}

func waitForAccountStatus(t *testing.T, svc *Service, accountID string, status Status) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		account, err := svc.Get(accountID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if account.Status == status {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	account, err := svc.Get(accountID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	t.Fatalf("Status = %q, want %q", account.Status, status)
}

func jsonRequest(method string, path string, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
