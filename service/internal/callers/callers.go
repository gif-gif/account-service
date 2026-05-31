package callers

import (
	"errors"
	"net/http"
	"sort"
	"strings"
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
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	APIKeyHash      string    `json:"-"`
	PlaintextAPIKey string    `json:"-"`
	Status          Status    `json:"status"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateResult struct {
	Caller          Caller `json:"caller"`
	PlaintextAPIKey string `json:"api_key"`
}

type CreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      Status `json:"status"`
}

type UpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *Status `json:"status"`
}

type Store interface {
	List() ([]Caller, error)
	CreateWithStatus(request CreateRequest) (CreateResult, error)
	Authenticate(apiKey string) (Caller, bool)
	RevealAPIKey(id string) (string, error)
	Update(id string, request UpdateRequest) (Caller, error)
	Delete(id string) error
}

type MemoryStore struct {
	mu      sync.Mutex
	callers map[string]Caller
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{callers: map[string]Caller{}}
}

func (store *MemoryStore) Create(name string, description string) (CreateResult, error) {
	return store.CreateWithStatus(CreateRequest{Name: name, Description: description, Status: StatusActive})
}

func (store *MemoryStore) CreateWithStatus(request CreateRequest) (CreateResult, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)
	if request.Status == "" {
		request.Status = StatusActive
	}
	if !validStatus(request.Status) {
		return CreateResult{}, errors.New("invalid caller status")
	}
	apiKey, err := security.GenerateAPIKey()
	if err != nil {
		return CreateResult{}, err
	}
	now := time.Now()
	caller := Caller{
		ID:              uuid.NewString(),
		Name:            request.Name,
		APIKeyHash:      security.HashAPIKey(apiKey),
		PlaintextAPIKey: apiKey,
		Status:          request.Status,
		Description:     request.Description,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	store.callers[caller.ID] = caller
	return CreateResult{Caller: caller, PlaintextAPIKey: apiKey}, nil
}

func (store *MemoryStore) List() ([]Caller, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	callers := make([]Caller, 0, len(store.callers))
	for _, caller := range store.callers {
		callers = append(callers, caller)
	}
	sort.SliceStable(callers, func(i, j int) bool {
		if callers[i].CreatedAt.Equal(callers[j].CreatedAt) {
			return callers[i].Name < callers[j].Name
		}
		return callers[i].CreatedAt.After(callers[j].CreatedAt)
	})
	return callers, nil
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

func (store *MemoryStore) RevealAPIKey(id string) (string, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	caller, ok := store.callers[id]
	if !ok {
		return "", errors.New("caller not found")
	}
	if caller.PlaintextAPIKey == "" {
		return "", errors.New("api key secret not available")
	}
	return caller.PlaintextAPIKey, nil
}

func (store *MemoryStore) Update(id string, request UpdateRequest) (Caller, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	caller, ok := store.callers[id]
	if !ok {
		return Caller{}, errors.New("caller not found")
	}
	if request.Name != nil {
		caller.Name = strings.TrimSpace(*request.Name)
	}
	if request.Description != nil {
		caller.Description = strings.TrimSpace(*request.Description)
	}
	if request.Status != nil {
		status := Status(strings.TrimSpace(string(*request.Status)))
		if !validStatus(status) {
			return Caller{}, errors.New("invalid caller status")
		}
		caller.Status = status
	}
	caller.UpdatedAt = time.Now()
	store.callers[id] = caller
	return caller, nil
}

func (store *MemoryStore) Disable(id string) error {
	status := StatusDisabled
	_, err := store.Update(id, UpdateRequest{Status: &status})
	return err
}

func (store *MemoryStore) Delete(id string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, ok := store.callers[id]; !ok {
		return errors.New("caller not found")
	}
	delete(store.callers, id)
	return nil
}

func RegisterRoutes(app *fiber.App, store Store) {
	app.Get("/api/v1/api-keys", func(c fiber.Ctx) error {
		callers, err := store.List()
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{"code": "internal_error", "message": "Failed to list API keys"}})
		}
		return c.JSON(fiber.Map{"callers": callers})
	})

	app.Post("/api/v1/api-keys", func(c fiber.Ctx) error {
		var request CreateRequest
		if err := c.Bind().Body(&request); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "invalid_request", "message": "Invalid API key request"}})
		}
		result, err := store.CreateWithStatus(request)
		if err != nil {
			return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"error": fiber.Map{"code": "invalid_api_key", "message": err.Error()}})
		}
		return c.Status(http.StatusCreated).JSON(result)
	})

	app.Get("/api/v1/api-keys/:id/secret", func(c fiber.Ctx) error {
		apiKey, err := store.RevealAPIKey(c.Params("id"))
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": fiber.Map{"code": "api_key_not_found", "message": "API key not found"}})
		}
		return c.JSON(fiber.Map{"api_key": apiKey})
	})

	app.Patch("/api/v1/api-keys/:id", func(c fiber.Ctx) error {
		var request UpdateRequest
		if err := c.Bind().Body(&request); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": fiber.Map{"code": "invalid_request", "message": "Invalid API key request"}})
		}
		caller, err := store.Update(c.Params("id"), request)
		if err != nil {
			return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"error": fiber.Map{"code": "invalid_api_key", "message": err.Error()}})
		}
		return c.JSON(fiber.Map{"caller": caller})
	})

	app.Delete("/api/v1/api-keys/:id", func(c fiber.Ctx) error {
		if err := store.Delete(c.Params("id")); err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": fiber.Map{"code": "api_key_not_found", "message": "API key not found"}})
		}
		return c.JSON(fiber.Map{"ok": true})
	})
}

func validStatus(status Status) bool {
	return status == StatusActive || status == StatusDisabled
}
