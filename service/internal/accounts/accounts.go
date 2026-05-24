package accounts

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"account-service/service/internal/audit"
	"account-service/service/internal/security"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Status string

const (
	StatusActive        Status = "active"
	StatusDisabled      Status = "disabled"
	StatusExhausted     Status = "exhausted"
	StatusLoginFailed   Status = "login_failed"
	StatusTokenExpired  Status = "token_expired"
	StatusRegionBlocked Status = "region_blocked"
	StatusError         Status = "error"
)

var validStatuses = map[Status]bool{
	StatusActive:        true,
	StatusDisabled:      true,
	StatusExhausted:     true,
	StatusLoginFailed:   true,
	StatusTokenExpired:  true,
	StatusRegionBlocked: true,
	StatusError:         true,
}

type Account struct {
	ID                  string   `json:"id"`
	Username            string   `json:"username"`
	Password            string   `json:"password,omitempty"`
	LoginURL            string   `json:"login_url"`
	AccessToken         string   `json:"access_token,omitempty"`
	RefreshToken        string   `json:"refresh_token,omitempty"`
	Region              string   `json:"region"`
	AccountType         string   `json:"account_type"`
	Status              Status   `json:"status"`
	QuotaTotal          int64    `json:"quota_total"`
	QuotaUsed           int64    `json:"quota_used"`
	QuotaRemaining      int64    `json:"quota_remaining"`
	MaxConcurrentLeases int      `json:"max_concurrent_leases"`
	Tags                []string `json:"tags"`
	Notes               string   `json:"notes"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type StoredAccount struct {
	Account
	PasswordEncrypted     string
	AccessTokenEncrypted  string
	RefreshTokenEncrypted string
}

type CreateAccountRequest struct {
	Username            string   `json:"username"`
	Password            string   `json:"password"`
	LoginURL            string   `json:"login_url"`
	AccessToken         string   `json:"access_token"`
	RefreshToken        string   `json:"refresh_token"`
	Region              string   `json:"region"`
	AccountType         string   `json:"account_type"`
	Status              Status   `json:"status"`
	QuotaTotal          int64    `json:"quota_total"`
	QuotaUsed           int64    `json:"quota_used"`
	QuotaRemaining      int64    `json:"quota_remaining"`
	MaxConcurrentLeases int      `json:"max_concurrent_leases"`
	Tags                []string `json:"tags"`
	Notes               string   `json:"notes"`
}

type UpdateAccountRequest struct {
	Username            *string  `json:"username"`
	Password            *string  `json:"password"`
	LoginURL            *string  `json:"login_url"`
	AccessToken         *string  `json:"access_token"`
	RefreshToken        *string  `json:"refresh_token"`
	Region              *string  `json:"region"`
	AccountType         *string  `json:"account_type"`
	Status              *Status  `json:"status"`
	QuotaTotal          *int64   `json:"quota_total"`
	QuotaUsed           *int64   `json:"quota_used"`
	QuotaRemaining      *int64   `json:"quota_remaining"`
	MaxConcurrentLeases *int     `json:"max_concurrent_leases"`
	Tags                []string `json:"tags"`
	Notes               *string  `json:"notes"`
}

type QueryRequest struct {
	Region            string   `json:"region"`
	AccountType       string   `json:"account_type"`
	Statuses          []Status `json:"statuses"`
	Tags              []string `json:"tags"`
	MinQuotaRemaining int64    `json:"min_quota_remaining"`
	Limit             int      `json:"limit"`
}

type MemoryRepository struct {
	mu       sync.Mutex
	accounts map[string]StoredAccount
	codec    security.CredentialCodec
}

func NewMemoryRepository(codec security.CredentialCodec) *MemoryRepository {
	return &MemoryRepository{accounts: map[string]StoredAccount{}, codec: codec}
}

func (repo *MemoryRepository) Raw(id string) (StoredAccount, bool) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	account, ok := repo.accounts[id]
	return account, ok
}

type Service struct {
	repo  *MemoryRepository
	codec security.CredentialCodec
	audit audit.Writer
}

func NewService(repo *MemoryRepository, codec security.CredentialCodec, auditWriter audit.Writer) *Service {
	return &Service{repo: repo, codec: codec, audit: auditWriter}
}

func (service *Service) Create(request CreateAccountRequest) (Account, error) {
	status := request.Status
	if status == "" {
		status = StatusActive
	}
	if !validStatuses[status] {
		return Account{}, errors.New("invalid account status")
	}
	if request.MaxConcurrentLeases <= 0 {
		request.MaxConcurrentLeases = 1
	}

	passwordEncrypted, err := service.codec.Encrypt(request.Password)
	if err != nil {
		return Account{}, err
	}
	accessTokenEncrypted, err := service.codec.Encrypt(request.AccessToken)
	if err != nil {
		return Account{}, err
	}
	refreshTokenEncrypted, err := service.codec.Encrypt(request.RefreshToken)
	if err != nil {
		return Account{}, err
	}

	now := time.Now()
	account := StoredAccount{
		Account: Account{
			ID:                  uuid.NewString(),
			Username:            request.Username,
			LoginURL:            request.LoginURL,
			Region:              request.Region,
			AccountType:         request.AccountType,
			Status:              status,
			QuotaTotal:          request.QuotaTotal,
			QuotaUsed:           request.QuotaUsed,
			QuotaRemaining:      request.QuotaRemaining,
			MaxConcurrentLeases: request.MaxConcurrentLeases,
			Tags:                normalizeTags(request.Tags),
			Notes:               request.Notes,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		PasswordEncrypted:     passwordEncrypted,
		AccessTokenEncrypted:  accessTokenEncrypted,
		RefreshTokenEncrypted: refreshTokenEncrypted,
	}

	service.repo.mu.Lock()
	service.repo.accounts[account.ID] = account
	service.repo.mu.Unlock()

	return service.decrypt(account)
}

func (service *Service) Get(id string) (Account, error) {
	service.repo.mu.Lock()
	account, ok := service.repo.accounts[id]
	service.repo.mu.Unlock()
	if !ok {
		return Account{}, errors.New("account not found")
	}
	return service.decrypt(account)
}

func (service *Service) Update(id string, request UpdateAccountRequest) (Account, error) {
	service.repo.mu.Lock()
	account, ok := service.repo.accounts[id]
	if !ok {
		service.repo.mu.Unlock()
		return Account{}, errors.New("account not found")
	}

	var err error
	if request.Username != nil {
		account.Username = *request.Username
	}
	if request.Password != nil {
		account.PasswordEncrypted, err = service.codec.Encrypt(*request.Password)
		if err != nil {
			service.repo.mu.Unlock()
			return Account{}, err
		}
	}
	if request.LoginURL != nil {
		account.LoginURL = *request.LoginURL
	}
	if request.AccessToken != nil {
		account.AccessTokenEncrypted, err = service.codec.Encrypt(*request.AccessToken)
		if err != nil {
			service.repo.mu.Unlock()
			return Account{}, err
		}
	}
	if request.RefreshToken != nil {
		account.RefreshTokenEncrypted, err = service.codec.Encrypt(*request.RefreshToken)
		if err != nil {
			service.repo.mu.Unlock()
			return Account{}, err
		}
	}
	if request.Region != nil {
		account.Region = *request.Region
	}
	if request.AccountType != nil {
		account.AccountType = *request.AccountType
	}
	if request.Status != nil {
		if !validStatuses[*request.Status] {
			service.repo.mu.Unlock()
			return Account{}, errors.New("invalid account status")
		}
		account.Status = *request.Status
	}
	if request.QuotaTotal != nil {
		account.QuotaTotal = *request.QuotaTotal
	}
	if request.QuotaUsed != nil {
		account.QuotaUsed = *request.QuotaUsed
	}
	if request.QuotaRemaining != nil {
		account.QuotaRemaining = *request.QuotaRemaining
	}
	if request.MaxConcurrentLeases != nil {
		account.MaxConcurrentLeases = *request.MaxConcurrentLeases
	}
	if request.Tags != nil {
		account.Tags = normalizeTags(request.Tags)
	}
	if request.Notes != nil {
		account.Notes = *request.Notes
	}
	account.UpdatedAt = time.Now()
	service.repo.accounts[id] = account
	service.repo.mu.Unlock()

	return service.decrypt(account)
}

func (service *Service) Query(request QueryRequest) ([]Account, error) {
	service.repo.mu.Lock()
	stored := make([]StoredAccount, 0, len(service.repo.accounts))
	for _, account := range service.repo.accounts {
		stored = append(stored, account)
	}
	service.repo.mu.Unlock()

	limit := request.Limit
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	out := make([]Account, 0, limit)
	for _, account := range stored {
		if request.Region != "" && account.Region != request.Region {
			continue
		}
		if request.AccountType != "" && account.AccountType != request.AccountType {
			continue
		}
		if len(request.Statuses) > 0 && !slices.Contains(request.Statuses, account.Status) {
			continue
		}
		if request.MinQuotaRemaining > 0 && account.QuotaRemaining < request.MinQuotaRemaining {
			continue
		}
		if !containsAllTags(account.Tags, request.Tags) {
			continue
		}
		decrypted, err := service.decrypt(account)
		if err != nil {
			return nil, err
		}
		out = append(out, decrypted)
		if len(out) == limit {
			break
		}
	}

	return out, nil
}

func (service *Service) decrypt(account StoredAccount) (Account, error) {
	out := account.Account
	var err error
	out.Password, err = service.codec.Decrypt(account.PasswordEncrypted)
	if err != nil {
		return Account{}, err
	}
	out.AccessToken, err = service.codec.Decrypt(account.AccessTokenEncrypted)
	if err != nil {
		return Account{}, err
	}
	out.RefreshToken, err = service.codec.Decrypt(account.RefreshTokenEncrypted)
	if err != nil {
		return Account{}, err
	}
	return out, nil
}

func RegisterRoutes(app *fiber.App, service *Service) {
	app.Post("/api/v1/accounts/query", service.handleQuery)
	app.Post("/api/v1/accounts", service.handleCreate)
	app.Get("/api/v1/accounts/:id", service.handleGet)
	app.Patch("/api/v1/accounts/:id", service.handleUpdate)
	app.Post("/api/v1/accounts/:id/status", service.handleStatus)
}

func (service *Service) handleQuery(c fiber.Ctx) error {
	var request QueryRequest
	if err := c.Bind().Body(&request); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid_request", "Invalid account query request")
	}
	accounts, err := service.Query(request)
	if err != nil {
		return jsonError(c, http.StatusInternalServerError, "internal_error", "Failed to query accounts")
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"accounts": accounts})
}

func (service *Service) handleCreate(c fiber.Ctx) error {
	var request CreateAccountRequest
	if err := c.Bind().Body(&request); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid_request", "Invalid account create request")
	}
	account, err := service.Create(request)
	if err != nil {
		return jsonError(c, http.StatusUnprocessableEntity, "invalid_account", err.Error())
	}
	c.Set("X-Test-Account-ID", account.ID)
	return c.Status(http.StatusCreated).JSON(fiber.Map{"account": account})
}

func (service *Service) handleGet(c fiber.Ctx) error {
	account, err := service.Get(c.Params("id"))
	if err != nil {
		return jsonError(c, http.StatusNotFound, "account_not_found", "Account not found")
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"account": account})
}

func (service *Service) handleUpdate(c fiber.Ctx) error {
	var request UpdateAccountRequest
	if err := c.Bind().Body(&request); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid_request", "Invalid account update request")
	}
	account, err := service.Update(c.Params("id"), request)
	if err != nil {
		return jsonError(c, http.StatusUnprocessableEntity, "invalid_account", err.Error())
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"account": account})
}

func (service *Service) handleStatus(c fiber.Ctx) error {
	var request struct {
		Status Status `json:"status"`
		Reason string `json:"reason"`
	}
	if err := c.Bind().Body(&request); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid_request", "Invalid status request")
	}
	account, err := service.Update(c.Params("id"), UpdateAccountRequest{Status: &request.Status})
	if err != nil {
		return jsonError(c, http.StatusUnprocessableEntity, "invalid_status", err.Error())
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{"account": account})
}

func jsonError(c fiber.Ctx, status int, code string, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": code, "message": message}})
}

func normalizeTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, tag := range tags {
		item := strings.TrimSpace(tag)
		if item != "" && !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}

func containsAllTags(haystack []string, needles []string) bool {
	for _, needle := range needles {
		if !slices.Contains(haystack, needle) {
			return false
		}
	}
	return true
}
