package leases

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"account-service/service/internal/accounts"
	"account-service/service/internal/audit"
	"account-service/service/internal/security"

	"github.com/gofiber/fiber/v3"
)

func TestAcquireSelectsHighestQuotaActiveAccount(t *testing.T) {
	accountService := newAccountService(t)
	low := createAccount(t, accountService, "low@example.com", 100, accounts.StatusActive, 1)
	high := createAccount(t, accountService, "high@example.com", 900, accounts.StatusActive, 1)
	_ = createAccount(t, accountService, "disabled@example.com", 1000, accounts.StatusDisabled, 1)

	service := NewService(accountService, 15*time.Minute, 2*time.Hour, audit.NewMemoryWriter())
	lease, err := service.Acquire(AcquireRequest{
		Region:            "us",
		AccountType:       string(accounts.AccountTypeCodex),
		Tags:              []string{"openai"},
		MinQuotaRemaining: 1,
		TTLSeconds:        900,
		Purpose:           "test",
		CallerID:          "caller-id",
	})
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if lease.Account.ID != high.ID {
		t.Fatalf("acquired account = %s, want high quota account %s; low was %s", lease.Account.ID, high.ID, low.ID)
	}
	if lease.ID == "" {
		t.Fatal("expected lease id")
	}
}

func TestAcquireRespectsMaxConcurrentLeases(t *testing.T) {
	accountService := newAccountService(t)
	account := createAccount(t, accountService, "one@example.com", 900, accounts.StatusActive, 1)
	service := NewService(accountService, 15*time.Minute, 2*time.Hour, audit.NewMemoryWriter())

	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.Acquire(AcquireRequest{Region: "us", AccountType: string(accounts.AccountTypeCodex), TTLSeconds: 900, CallerID: "caller-id"})
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	successes := 0
	failures := 0
	for err := range results {
		if err == nil {
			successes++
		} else {
			failures++
		}
	}
	if successes != 1 || failures != 1 {
		t.Fatalf("successes=%d failures=%d, want 1 and 1 for account %s", successes, failures, account.ID)
	}
}

func TestReleaseAndCleanupExpiredLease(t *testing.T) {
	accountService := newAccountService(t)
	createAccount(t, accountService, "one@example.com", 900, accounts.StatusActive, 1)
	service := NewService(accountService, 15*time.Minute, 2*time.Hour, audit.NewMemoryWriter())

	lease, err := service.Acquire(AcquireRequest{Region: "us", AccountType: string(accounts.AccountTypeCodex), TTLSeconds: 900, CallerID: "caller-id"})
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if err := service.Release(lease.ID); err != nil {
		t.Fatalf("Release() error = %v", err)
	}
	if err := service.Release(lease.ID); err == nil {
		t.Fatal("expected second release to fail")
	}

	expiring, err := service.Acquire(AcquireRequest{Region: "us", AccountType: string(accounts.AccountTypeCodex), TTLSeconds: 1, CallerID: "caller-id"})
	if err != nil {
		t.Fatalf("Acquire(expiring) error = %v", err)
	}
	service.ForceExpire(expiring.ID)
	expired := service.CleanupExpired(time.Now())
	if expired != 1 {
		t.Fatalf("CleanupExpired() = %d, want 1", expired)
	}
	if err := service.Release(expiring.ID); err == nil {
		t.Fatal("expected release of expired lease to fail")
	}
}

func TestAcquireRejectsTTLAboveMax(t *testing.T) {
	accountService := newAccountService(t)
	createAccount(t, accountService, "one@example.com", 900, accounts.StatusActive, 1)
	service := NewService(accountService, 15*time.Minute, 2*time.Hour, audit.NewMemoryWriter())

	_, err := service.Acquire(AcquireRequest{Region: "us", AccountType: string(accounts.AccountTypeCodex), TTLSeconds: int((3 * time.Hour).Seconds()), CallerID: "caller-id"})
	if err == nil {
		t.Fatal("expected max ttl error")
	}
}

func TestHandlersExposeLeaseAPI(t *testing.T) {
	accountService := newAccountService(t)
	createAccount(t, accountService, "one@example.com", 900, accounts.StatusActive, 1)
	service := NewService(accountService, 15*time.Minute, 2*time.Hour, audit.NewMemoryWriter())
	app := fiber.New()
	RegisterRoutes(app, service)

	acquireResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/accounts/acquire", `{"region":"us","account_type":"codex","ttl_seconds":900,"caller_id":"caller-id"}`))
	if err != nil {
		t.Fatalf("acquire app.Test() error = %v", err)
	}
	if acquireResp.StatusCode != http.StatusOK {
		t.Fatalf("acquire status = %d, want %d", acquireResp.StatusCode, http.StatusOK)
	}
	leaseID := acquireResp.Header.Get("X-Test-Lease-ID")
	if leaseID == "" {
		t.Fatal("expected X-Test-Lease-ID header")
	}

	releaseResp, err := app.Test(jsonRequest(http.MethodPost, "/api/v1/accounts/release", `{"lease_id":"`+leaseID+`"}`))
	if err != nil {
		t.Fatalf("release app.Test() error = %v", err)
	}
	if releaseResp.StatusCode != http.StatusOK {
		t.Fatalf("release status = %d, want %d", releaseResp.StatusCode, http.StatusOK)
	}

	listResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/leases?status=released", nil))
	if err != nil {
		t.Fatalf("list app.Test() error = %v", err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listResp.StatusCode, http.StatusOK)
	}
}

func newAccountService(t *testing.T) *accounts.Service {
	t.Helper()
	codec, err := security.NewCredentialCodec("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("NewCredentialCodec() error = %v", err)
	}
	return accounts.NewService(accounts.NewMemoryRepository(codec), codec, audit.NewMemoryWriter())
}

func createAccount(t *testing.T, service *accounts.Service, username string, quota int64, status accounts.Status, maxLeases int) accounts.Account {
	t.Helper()
	account, err := service.Create(accounts.CreateAccountRequest{
		Username:            username,
		Password:            "plain-password",
		LoginURL:            "https://example.com/login",
		AccessToken:         "access-token",
		RefreshToken:        "refresh-token",
		Region:              "us",
		AccountType:         accounts.AccountTypeCodex,
		Status:              status,
		QuotaTotal:          1000,
		QuotaRemaining:      quota,
		MaxConcurrentLeases: maxLeases,
		Tags:                []string{"openai"},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	return account
}

func jsonRequest(method string, path string, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
