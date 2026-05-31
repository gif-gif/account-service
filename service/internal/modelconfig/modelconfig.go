package modelconfig

import (
	"maps"
	"sync"

	"github.com/gofiber/fiber/v3"
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

var FallbackModels = []Model{
	{ModelID: "auto"},
	{ModelID: "claude-sonnet-4"},
	{ModelID: "claude-haiku-4.5"},
	{ModelID: "claude-sonnet-4.5"},
	{ModelID: "claude-opus-4.5"},
	{ModelID: "claude-opus-4.6"},
	{ModelID: "claude-sonnet-4.6"},
}

var HiddenModels = map[string]string{
	"claude-3.7-sonnet": "CLAUDE_3_7_SONNET_20250219_V1_0",
	"claude-opus-4.6":   "claude-opus-4.6",
	"claude-sonnet-4.6": "claude-sonnet-4.6",
}

var ModelAliases = map[string]string{
	"auto-kiro":         "auto",
	"claude-opus-4-6":   "claude-opus-4.6",
	"claude-sonnet-4-6": "claude-sonnet-4.6",
	"claude-opus-4-5":   "claude-opus-4.5",
	"claude-sonnet-4-5": "claude-sonnet-4.5",
	"claude-haiku-4-5":  "claude-haiku-4.5",
	"claude-opus-4-7":   "claude-opus-4.7",
}

var HiddenFromList = []string{"auto"}

type Service struct {
	mu     sync.RWMutex
	config Config
}

func NewDefaultService() *Service {
	return NewService(Config{
		FallbackModels: FallbackModels,
		HiddenModels:   HiddenModels,
		ModelAliases:   ModelAliases,
		HiddenFromList: HiddenFromList,
	})
}

func NewService(config Config) *Service {
	return &Service{config: cloneConfig(config)}
}

func (service *Service) Get() Config {
	service.mu.RLock()
	defer service.mu.RUnlock()
	return cloneConfig(service.config)
}

func RegisterExternalRoutes(app *fiber.App, service *Service) {
	app.Get("/api/v1/external/model-config", func(c fiber.Ctx) error {
		return c.JSON(service.Get())
	})
}

func cloneConfig(config Config) Config {
	return Config{
		FallbackModels: append([]Model(nil), config.FallbackModels...),
		HiddenModels:   maps.Clone(config.HiddenModels),
		ModelAliases:   maps.Clone(config.ModelAliases),
		HiddenFromList: append([]string(nil), config.HiddenFromList...),
	}
}
