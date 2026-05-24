package callers

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"account-service/service/internal/security"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

type Caller struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	APIKeyHash  string    `json:"-"`
	Status      Status    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateResult struct {
	Caller          Caller `json:"caller"`
	PlaintextAPIKey string `json:"api_key"`
}

type MemoryStore struct {
	mu      sync.Mutex
	callers map[string]Caller
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{callers: map[string]Caller{}}
}

func (store *MemoryStore) Create(name string, description string) (CreateResult, error) {
	apiKey, err := security.GenerateAPIKey()
	if err != nil {
		return CreateResult{}, err
	}
	now := time.Now()
	caller := Caller{
		ID:          uuid.NewString(),
		Name:        name,
		APIKeyHash:  security.HashAPIKey(apiKey),
		Status:      StatusActive,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	store.callers[caller.ID] = caller
	return CreateResult{Caller: caller, PlaintextAPIKey: apiKey}, nil
}

func (store *MemoryStore) Authenticate(apiKey string) (Caller, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()

	for _, caller := range store.callers {
		if caller.Status == StatusActive && security.VerifyAPIKey(apiKey, caller.APIKeyHash) {
			return caller, true
		}
	}
	return Caller{}, false
}

func (store *MemoryStore) Disable(id string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	caller, ok := store.callers[id]
	if !ok {
		return errors.New("caller not found")
	}
	caller.Status = StatusDisabled
	caller.UpdatedAt = time.Now()
	store.callers[id] = caller
	return nil
}

func RegisterRoutes(app *fiber.App, store *MemoryStore) {
	app.Post("/api/v1/api-keys", func(c fiber.Ctx) error {
		var request struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := c.Bind().Body(&request); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "invalid_request", "message": "Invalid API key request"}})
		}
		result, err := store.Create(request.Name, request.Description)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"code": "internal_error", "message": "Failed to create API key"}})
		}
		return c.Status(http.StatusCreated).JSON(result)
	})
}
