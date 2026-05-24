package app

import (
	"context"

	"account-service/service/internal/accounts"
	"account-service/service/internal/callers"
	"account-service/service/internal/health"
	"account-service/service/internal/httpx"
	"account-service/service/internal/leases"

	"github.com/gofiber/fiber/v3"
)

type Options struct {
	HealthChecker  health.Checker
	AccountService *accounts.Service
	LeaseService   *leases.Service
	CallerStore    *callers.MemoryStore
	CORSOrigins    []string
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
	if options.AccountService != nil {
		accounts.RegisterRoutes(fiberApp, options.AccountService)
	}
	if options.LeaseService != nil {
		leases.RegisterRoutes(fiberApp, options.LeaseService)
	}
	if options.CallerStore != nil {
		callers.RegisterRoutes(fiberApp, options.CallerStore)
	}

	return fiberApp
}
