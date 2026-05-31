package modelconfig

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"account-service/service/internal/httpx"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Kind string

const (
	KindFallbackModel  Kind = "fallback_model"
	KindHiddenModel    Kind = "hidden_model"
	KindModelAlias     Kind = "model_alias"
	KindHiddenFromList Kind = "hidden_from_list"
)

type Model struct {
	ModelID string `json:"model_id"`
}

type Config struct {
	FallbackModels []Model           `json:"fallback_models"`
	HiddenModels   map[string]string `json:"hidden_models"`
	ModelAliases   map[string]string `json:"model_aliases"`
	HiddenFromList []string          `json:"hidden_from_list"`
}

type Item struct {
	ID           string    `json:"id"`
	Kind         Kind      `json:"kind"`
	Key          string    `json:"key"`
	Value        string    `json:"value"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateItemRequest struct {
	Kind         Kind   `json:"kind"`
	Key          string `json:"key"`
	Value        string `json:"value"`
	DisplayOrder int    `json:"display_order"`
}

type UpdateItemRequest struct {
	Kind         *Kind   `json:"kind"`
	Key          *string `json:"key"`
	Value        *string `json:"value"`
	DisplayOrder *int    `json:"display_order"`
}

type Repository interface {
	List(ctx context.Context) ([]Item, error)
	Create(ctx context.Context, request CreateItemRequest) (Item, error)
	Update(ctx context.Context, id string, request UpdateItemRequest) (Item, error)
	Delete(ctx context.Context, id string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (service *Service) Config(ctx context.Context) (Config, error) {
	items, err := service.repo.List(ctx)
	if err != nil {
		return Config{}, err
	}
	return configFromItems(items), nil
}

func (service *Service) List(ctx context.Context) ([]Item, error) {
	return service.repo.List(ctx)
}

func (service *Service) Create(ctx context.Context, request CreateItemRequest) (Item, error) {
	normalized, err := normalizeCreateRequest(request)
	if err != nil {
		return Item{}, err
	}
	return service.repo.Create(ctx, normalized)
}

func (service *Service) Update(ctx context.Context, id string, request UpdateItemRequest) (Item, error) {
	normalized, err := normalizeUpdateRequest(request)
	if err != nil {
		return Item{}, err
	}
	return service.repo.Update(ctx, id, normalized)
}

func (service *Service) Delete(ctx context.Context, id string) error {
	return service.repo.Delete(ctx, id)
}

type MemoryRepository struct {
	mu    sync.Mutex
	items map[string]Item
}

func NewMemoryRepository(items []Item) *MemoryRepository {
	repo := &MemoryRepository{items: map[string]Item{}}
	for _, item := range items {
		if item.ID == "" {
			item.ID = uuid.NewString()
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = time.Now()
		}
		if item.UpdatedAt.IsZero() {
			item.UpdatedAt = item.CreatedAt
		}
		repo.items[item.ID] = item
	}
	return repo
}

func (repo *MemoryRepository) List(ctx context.Context) ([]Item, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	items := make([]Item, 0, len(repo.items))
	for _, item := range repo.items {
		items = append(items, item)
	}
	sortItems(items)
	return items, nil
}

func (repo *MemoryRepository) Create(ctx context.Context, request CreateItemRequest) (Item, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	item := Item{
		ID:           uuid.NewString(),
		Kind:         request.Kind,
		Key:          request.Key,
		Value:        request.Value,
		DisplayOrder: request.DisplayOrder,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.items[item.ID] = item
	return item, nil
}

func (repo *MemoryRepository) Update(ctx context.Context, id string, request UpdateItemRequest) (Item, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	item, ok := repo.items[id]
	if !ok {
		return Item{}, errors.New("model config item not found")
	}
	if request.Kind != nil {
		item.Kind = *request.Kind
	}
	if request.Key != nil {
		item.Key = *request.Key
	}
	if request.Value != nil {
		item.Value = *request.Value
	}
	if request.DisplayOrder != nil {
		item.DisplayOrder = *request.DisplayOrder
	}
	item.UpdatedAt = time.Now()
	repo.items[id] = item
	return item, nil
}

func (repo *MemoryRepository) Delete(ctx context.Context, id string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, ok := repo.items[id]; !ok {
		return errors.New("model config item not found")
	}
	delete(repo.items, id)
	return nil
}

func RegisterRoutes(app *fiber.App, service *Service) {
	app.Get("/api/v1/model-config/items", service.handleList)
	app.Post("/api/v1/model-config/items", service.handleCreate)
	app.Patch("/api/v1/model-config/items/:id", service.handleUpdate)
	app.Delete("/api/v1/model-config/items/:id", service.handleDelete)
}

func RegisterExternalRoutes(app *fiber.App, service *Service) {
	app.Get("/api/v1/external/model-config", service.handleExternalConfig)
}

func (service *Service) handleExternalConfig(c fiber.Ctx) error {
	config, err := service.Config(c.Context())
	if err != nil {
		return httpx.JSONError(c, http.StatusInternalServerError, "internal_error", "Failed to load model config")
	}
	return c.JSON(config)
}

func (service *Service) handleList(c fiber.Ctx) error {
	items, err := service.List(c.Context())
	if err != nil {
		return httpx.JSONError(c, http.StatusInternalServerError, "internal_error", "Failed to list model config items")
	}
	return c.JSON(fiber.Map{"items": items})
}

func (service *Service) handleCreate(c fiber.Ctx) error {
	var request CreateItemRequest
	if err := c.Bind().Body(&request); err != nil {
		return httpx.JSONError(c, http.StatusBadRequest, "invalid_request", "Invalid model config item request")
	}
	item, err := service.Create(c.Context(), request)
	if err != nil {
		return httpx.JSONError(c, http.StatusUnprocessableEntity, "invalid_model_config_item", err.Error())
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"item": item})
}

func (service *Service) handleUpdate(c fiber.Ctx) error {
	var request UpdateItemRequest
	if err := c.Bind().Body(&request); err != nil {
		return httpx.JSONError(c, http.StatusBadRequest, "invalid_request", "Invalid model config item request")
	}
	item, err := service.Update(c.Context(), c.Params("id"), request)
	if err != nil {
		return httpx.JSONError(c, http.StatusUnprocessableEntity, "invalid_model_config_item", err.Error())
	}
	return c.JSON(fiber.Map{"item": item})
}

func (service *Service) handleDelete(c fiber.Ctx) error {
	if err := service.Delete(c.Context(), c.Params("id")); err != nil {
		return httpx.JSONError(c, http.StatusNotFound, "model_config_item_not_found", "Model config item not found")
	}
	return c.JSON(fiber.Map{"ok": true})
}

func configFromItems(items []Item) Config {
	sortItems(items)
	config := Config{
		FallbackModels: []Model{},
		HiddenModels:   map[string]string{},
		ModelAliases:   map[string]string{},
		HiddenFromList: []string{},
	}
	for _, item := range items {
		switch item.Kind {
		case KindFallbackModel:
			config.FallbackModels = append(config.FallbackModels, Model{ModelID: item.Key})
		case KindHiddenModel:
			config.HiddenModels[item.Key] = item.Value
		case KindModelAlias:
			config.ModelAliases[item.Key] = item.Value
		case KindHiddenFromList:
			config.HiddenFromList = append(config.HiddenFromList, item.Key)
		}
	}
	return config
}

func sortItems(items []Item) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Kind != items[j].Kind {
			return items[i].Kind < items[j].Kind
		}
		if items[i].DisplayOrder != items[j].DisplayOrder {
			return items[i].DisplayOrder < items[j].DisplayOrder
		}
		return items[i].Key < items[j].Key
	})
}

func normalizeCreateRequest(request CreateItemRequest) (CreateItemRequest, error) {
	request.Kind = Kind(strings.TrimSpace(string(request.Kind)))
	request.Key = strings.TrimSpace(request.Key)
	request.Value = strings.TrimSpace(request.Value)
	if err := validateItemFields(request.Kind, request.Key, request.Value); err != nil {
		return CreateItemRequest{}, err
	}
	return request, nil
}

func normalizeUpdateRequest(request UpdateItemRequest) (UpdateItemRequest, error) {
	if request.Kind != nil {
		kind := Kind(strings.TrimSpace(string(*request.Kind)))
		request.Kind = &kind
	}
	if request.Key != nil {
		key := strings.TrimSpace(*request.Key)
		request.Key = &key
	}
	if request.Value != nil {
		value := strings.TrimSpace(*request.Value)
		request.Value = &value
	}
	if request.Kind != nil && !validKind(*request.Kind) {
		return UpdateItemRequest{}, errors.New("invalid model config kind")
	}
	if request.Key != nil && *request.Key == "" {
		return UpdateItemRequest{}, errors.New("key is required")
	}
	return request, nil
}

func validateItemFields(kind Kind, key string, value string) error {
	if !validKind(kind) {
		return errors.New("invalid model config kind")
	}
	if key == "" {
		return errors.New("key is required")
	}
	if (kind == KindHiddenModel || kind == KindModelAlias) && value == "" {
		return errors.New("value is required")
	}
	return nil
}

func validKind(kind Kind) bool {
	return kind == KindFallbackModel ||
		kind == KindHiddenModel ||
		kind == KindModelAlias ||
		kind == KindHiddenFromList
}
