package modelconfig

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestDefaultServiceReturnsModelConfiguration(t *testing.T) {
	service := NewDefaultService()

	config := service.Get()
	if len(config.FallbackModels) != 7 {
		t.Fatalf("FallbackModels len = %d, want 7", len(config.FallbackModels))
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

	config.FallbackModels[0].ModelID = "mutated"
	config.HiddenModels["claude-3.7-sonnet"] = "mutated"
	config.ModelAliases["auto-kiro"] = "mutated"
	config.HiddenFromList[0] = "mutated"

	fresh := service.Get()
	if fresh.FallbackModels[0].ModelID != "auto" ||
		fresh.HiddenModels["claude-3.7-sonnet"] != "CLAUDE_3_7_SONNET_20250219_V1_0" ||
		fresh.ModelAliases["auto-kiro"] != "auto" ||
		fresh.HiddenFromList[0] != "auto" {
		t.Fatalf("Get() returned mutable internal config: %#v", fresh)
	}
}

func TestRegisterExternalRoutesReturnsUnifiedModelConfig(t *testing.T) {
	app := fiber.New()
	RegisterExternalRoutes(app, NewDefaultService())

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/external/model-config", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body Config
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.FallbackModels[0].ModelID != "auto" {
		t.Fatalf("fallback_models[0].model_id = %q, want auto", body.FallbackModels[0].ModelID)
	}
	if body.HiddenModels["claude-opus-4.6"] != "claude-opus-4.6" {
		t.Fatalf("hidden_models missing claude-opus-4.6: %#v", body.HiddenModels)
	}
	if body.ModelAliases["claude-sonnet-4-6"] != "claude-sonnet-4.6" {
		t.Fatalf("model_aliases missing claude-sonnet-4-6: %#v", body.ModelAliases)
	}
	if len(body.HiddenFromList) != 1 || body.HiddenFromList[0] != "auto" {
		t.Fatalf("hidden_from_list = %#v, want [auto]", body.HiddenFromList)
	}
}
