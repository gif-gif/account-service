package leases

import (
	"errors"
	"net/http"
	"sort"
	"sync"
	"time"

	"account-service/service/internal/accounts"
	"account-service/service/internal/audit"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusReleased Status = "released"
	StatusExpired  Status = "expired"
)

type AcquireRequest struct {
	Region            string   `json:"region"`
	AccountType       string   `json:"account_type"`
	Tags              []string `json:"tags"`
	MinQuotaRemaining int64    `json:"min_quota_remaining"`
	TTLSeconds        int      `json:"ttl_seconds"`
	Purpose           string   `json:"purpose"`
	CallerID          string   `json:"caller_id"`
}

type ReleaseRequest struct {
	LeaseID string `json:"lease_id"`
}

type Lease struct {
	ID             string           `json:"lease_id"`
	AccountID      string           `json:"account_id"`
	CallerID       string           `json:"caller_id"`
	Purpose        string           `json:"purpose"`
	Status         Status           `json:"status"`
	LeasedAt       time.Time        `json:"leased_at"`
	ExpiresAt      time.Time        `json:"expires_at"`
	ReleasedAt     *time.Time       `json:"released_at,omitempty"`
	RequestFilters AcquireRequest   `json:"request_filters"`
	Account        accounts.Account `json:"account,omitempty"`
}

type Service struct {
	mu         sync.Mutex
	accounts   *accounts.Service
	defaultTTL time.Duration
	maxTTL     time.Duration
	audit      audit.Writer
	leases     map[string]Lease
}

func NewService(accountService *accounts.Service, defaultTTL time.Duration, maxTTL time.Duration, auditWriter audit.Writer) *Service {
	return &Service{
		accounts:   accountService,
		defaultTTL: defaultTTL,
		maxTTL:     maxTTL,
		audit:      auditWriter,
		leases:     map[string]Lease{},
	}
}

func (service *Service) Acquire(request AcquireRequest) (Lease, error) {
	ttl := service.defaultTTL
	if request.TTLSeconds > 0 {
		ttl = time.Duration(request.TTLSeconds) * time.Second
	}
	if ttl > service.maxTTL {
		return Lease{}, errors.New("ttl exceeds maximum")
	}

	candidates, err := service.accounts.Query(accounts.QueryRequest{
		Region:            request.Region,
		AccountType:       accounts.AccountType(request.AccountType),
		Statuses:          []accounts.Status{accounts.StatusActive},
		Tags:              request.Tags,
		MinQuotaRemaining: maxInt64(request.MinQuotaRemaining, 1),
		Limit:             100,
	})
	if err != nil {
		return Lease{}, err
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].QuotaRemaining > candidates[j].QuotaRemaining
	})

	service.mu.Lock()
	defer service.mu.Unlock()

	now := time.Now()
	for _, candidate := range candidates {
		if service.activeLeaseCountLocked(candidate.ID, now) >= candidate.MaxConcurrentLeases {
			continue
		}
		lease := Lease{
			ID:             uuid.NewString(),
			AccountID:      candidate.ID,
			CallerID:       request.CallerID,
			Purpose:        request.Purpose,
			Status:         StatusActive,
			LeasedAt:       now,
			ExpiresAt:      now.Add(ttl),
			RequestFilters: request,
			Account:        candidate,
		}
		service.leases[lease.ID] = lease
		return lease, nil
	}

	return Lease{}, errors.New("no available account")
}

func (service *Service) Release(leaseID string) error {
	service.mu.Lock()
	defer service.mu.Unlock()

	lease, ok := service.leases[leaseID]
	if !ok {
		return errors.New("lease not found")
	}
	if lease.Status != StatusActive || time.Now().After(lease.ExpiresAt) {
		return errors.New("lease is not active")
	}

	now := time.Now()
	lease.Status = StatusReleased
	lease.ReleasedAt = &now
	service.leases[lease.ID] = lease
	return nil
}

func (service *Service) CleanupExpired(now time.Time) int {
	service.mu.Lock()
	defer service.mu.Unlock()

	expired := 0
	for id, lease := range service.leases {
		if lease.Status == StatusActive && !now.Before(lease.ExpiresAt) {
			lease.Status = StatusExpired
			service.leases[id] = lease
			expired++
		}
	}
	return expired
}

func (service *Service) ForceExpire(leaseID string) {
	service.mu.Lock()
	defer service.mu.Unlock()

	lease, ok := service.leases[leaseID]
	if !ok {
		return
	}
	lease.ExpiresAt = time.Now().Add(-time.Second)
	service.leases[leaseID] = lease
}

func (service *Service) List(status Status) []Lease {
	service.mu.Lock()
	defer service.mu.Unlock()

	out := make([]Lease, 0, len(service.leases))
	for _, lease := range service.leases {
		if status == "" || lease.Status == status {
			out = append(out, lease)
		}
	}
	return out
}

func (service *Service) activeLeaseCountLocked(accountID string, now time.Time) int {
	count := 0
	for _, lease := range service.leases {
		if lease.AccountID == accountID && lease.Status == StatusActive && now.Before(lease.ExpiresAt) {
			count++
		}
	}
	return count
}

func RegisterRoutes(app *fiber.App, service *Service) {
	app.Post("/api/v1/accounts/acquire", service.handleAcquire)
	app.Post("/api/v1/accounts/release", service.handleRelease)
	app.Get("/api/v1/leases", service.handleList)
}

func (service *Service) handleAcquire(c fiber.Ctx) error {
	var request AcquireRequest
	if err := c.Bind().Body(&request); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid_request", "Invalid acquire request")
	}
	lease, err := service.Acquire(request)
	if err != nil {
		return jsonError(c, http.StatusNotFound, "no_available_account", err.Error())
	}
	c.Set("X-Test-Lease-ID", lease.ID)
	return c.Status(http.StatusOK).JSON(lease)
}

func (service *Service) handleRelease(c fiber.Ctx) error {
	var request ReleaseRequest
	if err := c.Bind().Body(&request); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid_request", "Invalid release request")
	}
	if err := service.Release(request.LeaseID); err != nil {
		return jsonError(c, http.StatusConflict, "lease_conflict", err.Error())
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"ok": true})
}

func (service *Service) handleList(c fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{"leases": service.List(Status(c.Query("status")))})
}

func jsonError(c fiber.Ctx, status int, code string, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": code, "message": message}})
}

func maxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
