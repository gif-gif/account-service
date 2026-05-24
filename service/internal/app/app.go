package app

import (
	"context"

	"account-service/service/internal/accounts"
	"account-service/service/internal/health"

	"github.com/gofiber/fiber/v3"
)

type Options struct {
	HealthChecker  health.Checker
	AccountService *accounts.Service
}

func New(options Options) *fiber.App {
	fiberApp := fiber.New(fiber.Config{
		AppName: "account-service",
	})

	checker := options.HealthChecker
	if checker == nil {
		checker = health.CheckerFunc(func(context.Context) error { return nil })
	}
	health.Register(fiberApp, checker)
	if options.AccountService != nil {
		accounts.RegisterRoutes(fiberApp, options.AccountService)
	}

	return fiberApp
}
