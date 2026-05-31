package app

import (
	"context"

	"account-service/service/internal/accounts"
	"account-service/service/internal/admin"
	"account-service/service/internal/callers"
	"account-service/service/internal/health"
	"account-service/service/internal/httpx"
	"account-service/service/internal/leases"
	"account-service/service/internal/modelconfig"

	"github.com/gofiber/fiber/v3"
)

type Options struct {
	HealthChecker             health.Checker
	AdminService              *admin.Service
	AccountService            *accounts.Service
	LeaseService              *leases.Service
	CallerStore               *callers.MemoryStore
	ModelConfig               *modelconfig.Service
	ExternalAPIKeyAuthEnabled *bool
	CORSOrigins               []string
}

func New(options Options) *fiber.App {
	fiberApp := fiber.New(fiber.Config{
		AppName: "account-service",
	})
	fiberApp.Use(httpx.RequestID())
	if len(options.CORSOrigins) > 0 {
		fiberApp.Use(httpx.CORS(options.CORSOrigins))
	}

	checker := options.HealthChecker
	if checker == nil {
		checker = health.CheckerFunc(func(context.Context) error { return nil })
	}
	health.Register(fiberApp, checker)
	if options.AdminService != nil {
		admin.RegisterRoutes(fiberApp, options.AdminService)
		fiberApp.Use("/api/v1/accounts", httpx.AdminSession(options.AdminService))
		fiberApp.Use("/api/v1/leases", httpx.AdminSession(options.AdminService))
		fiberApp.Use("/api/v1/api-keys", httpx.AdminSession(options.AdminService))
		fiberApp.Use("/api/v1/model-config", httpx.AdminSession(options.AdminService))
	}
	if options.AccountService != nil {
		accounts.RegisterRoutes(fiberApp, options.AccountService)
	}
	if options.LeaseService != nil {
		leases.RegisterRoutes(fiberApp, options.LeaseService)
	}
	modelConfig := options.ModelConfig
	if modelConfig != nil {
		modelconfig.RegisterRoutes(fiberApp, modelConfig)
	}
	if options.CallerStore != nil {
		callers.RegisterRoutes(fiberApp, options.CallerStore)
		externalAPIKeyAuthEnabled := true
		if options.ExternalAPIKeyAuthEnabled != nil {
			externalAPIKeyAuthEnabled = *options.ExternalAPIKeyAuthEnabled
		}
		if externalAPIKeyAuthEnabled {
			fiberApp.Use("/api/v1/external", httpx.APIKeyAuth(options.CallerStore))
		}
		if options.AccountService != nil {
			accounts.RegisterExternalRoutes(fiberApp, options.AccountService)
		}
		if options.LeaseService != nil {
			leases.RegisterExternalRoutes(fiberApp, options.LeaseService)
		}
		if modelConfig == nil {
			modelConfig = modelconfig.NewService(modelconfig.NewMemoryRepository(nil))
		}
		modelconfig.RegisterExternalRoutes(fiberApp, modelConfig)
	}

	return fiberApp
}
