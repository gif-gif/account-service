package modelconfig

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"account-service/service/internal/db"
	"account-service/service/internal/testutil"

	"github.com/gofiber/fiber/v3"
)

func TestServiceBuildsUnifiedConfigFromRepositoryItems(t *testing.T) {
	service := NewService(NewMemoryRepository([]Item{
		{Kind: KindFallbackModel, Key: "claude-sonnet-4.6", DisplayOrder: 20},
		{Kind: KindFallbackModel, Key: "auto", DisplayOrder: 10},
		{Kind: KindHiddenModel, Key: "claude-3.7-sonnet", Value: "CLAUDE_3_7_SONNET_20250219_V1_0"},
		{Kind: KindModelAlias, Key: "claude-opus-4-7", Value: "claude-opus-4.7"},
		{Kind: KindHiddenFromList, Key: "auto"},
	}))

	config, err := service.Config(context.Background())
	if err != nil {
		t.Fatalf("Config() error = %v", err)
	}
	if len(config.FallbackModels) != 2 {
		t.Fatalf("FallbackModels len = %d, want 2", len(config.FallbackModels))
	}
	if config.FallbackModels[0].ModelID != "auto" {
		t.Fatalf("first fallback model = %q, want auto", config.FallbackModels[0].ModelID)
	}
	if config.HiddenModels["claude-3.7-sonnet"] != "CLAUDE_3_7_SONNET_20250219_V1_0" {
		t.Fatalf("hidden claude-3.7-sonnet = %q", config.HiddenModels["claude-3.7-sonnet"])
	}
	if config.ModelAliases["claude-opus-4-7"] != "claude-opus-4.7" {
		t.Fatalf("alias claude-opus-4-7 = %q", config.ModelAliases["claude-opus-4-7"])
	}
	if len(config.HiddenFromList) != 1 || config.HiddenFromList[0] != "auto" {
		t.Fatalf("HiddenFromList = %#v, want [auto]", config.HiddenFromList)
	}
}

func TestServiceValidatesAndMutatesItems(t *testing.T) {
	service := NewService(NewMemoryRepository(nil))

	_, err := service.Create(context.Background(), CreateItemRequest{Kind: KindModelAlias, Key: "alias"})
	if err == nil {
		t.Fatal("expected model alias without value to fail")
	}

	created, err := service.Create(context.Background(), CreateItemRequest{
		Kind:         KindModelAlias,
		Key:          " alias ",
		Value:        " target ",
		DisplayOrder: 10,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Key != "alias" || created.Value != "target" {
		t.Fatalf("created item was not normalized: %#v", created)
	}

	nextValue := "new-target"
	updated, err := service.Update(context.Background(), created.ID, UpdateItemRequest{Value: &nextValue})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Value != "new-target" {
		t.Fatalf("updated value = %q, want new-target", updated.Value)
	}
	if err := service.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	items, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items len = %d, want 0", len(items))
	}
}

func TestRegisterRoutesExposeAdminCRUDAndExternalConfig(t *testing.T) {
	service := NewService(NewMemoryRepository([]Item{{Kind: KindFallbackModel, Key: "auto", DisplayOrder: 10}}))
	app := fiber.New()
	RegisterRoutes(app, service)
	RegisterExternalRoutes(app, service)

	listResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/model-config/items", nil))
	if err != nil {
		t.Fatalf("list app.Test() error = %v", err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listResp.StatusCode, http.StatusOK)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/model-config/items", stringsReader(`{
		"kind": "model_alias",
		"key": "claude-opus-4-7",
		"value": "claude-opus-4.7",
		"display_order": 20
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createResp, err := app.Test(createReq)
	if err != nil {
		t.Fatalf("create app.Test() error = %v", err)
	}
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResp.StatusCode, http.StatusCreated)
	}
	var createBody struct {
		Item Item `json:"item"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&createBody); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/model-config/items/"+createBody.Item.ID, stringsReader(`{"value":"claude-opus-4.8"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp, err := app.Test(updateReq)
	if err != nil {
		t.Fatalf("update app.Test() error = %v", err)
	}
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d, want %d", updateResp.StatusCode, http.StatusOK)
	}

	configResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/external/model-config", nil))
	if err != nil {
		t.Fatalf("config app.Test() error = %v", err)
	}
	var config Config
	if err := json.NewDecoder(configResp.Body).Decode(&config); err != nil {
		t.Fatalf("decode config response: %v", err)
	}
	if config.ModelAliases["claude-opus-4-7"] != "claude-opus-4.8" {
		t.Fatalf("model alias = %q, want claude-opus-4.8", config.ModelAliases["claude-opus-4-7"])
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/model-config/items/"+createBody.Item.ID, nil)
	deleteResp, err := app.Test(deleteReq)
	if err != nil {
		t.Fatalf("delete app.Test() error = %v", err)
	}
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("delete status = %d, want %d", deleteResp.StatusCode, http.StatusOK)
	}
}

func TestPostgresRepositoryPersistsModelConfigItems(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool := testutil.OpenTestDB(t, ctx, databaseURL)
	testutil.ResetSchema(t, ctx, pool)
	if err := db.ApplyMigrations(ctx, pool); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	service := NewService(NewPostgresRepository(pool))
	config, err := service.Config(ctx)
	if err != nil {
		t.Fatalf("Config() error = %v", err)
	}
	if len(config.FallbackModels) == 0 || config.FallbackModels[0].ModelID != "auto" {
		t.Fatalf("seeded fallback models = %#v, want first auto", config.FallbackModels)
	}

	created, err := service.Create(ctx, CreateItemRequest{Kind: KindModelAlias, Key: "test-alias", Value: "test-target", DisplayOrder: 99})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	nextValue := "test-target-2"
	if _, err := service.Update(ctx, created.ID, UpdateItemRequest{Value: &nextValue}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	config, err = service.Config(ctx)
	if err != nil {
		t.Fatalf("Config() after update error = %v", err)
	}
	if config.ModelAliases["test-alias"] != "test-target-2" {
		t.Fatalf("test-alias = %q, want test-target-2", config.ModelAliases["test-alias"])
	}
	if err := service.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func stringsReader(value string) *strings.Reader {
	return strings.NewReader(value)
}
