package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"account-service/service/internal/accounts"
	"account-service/service/internal/admin"
	"account-service/service/internal/app"
	"account-service/service/internal/audit"
	"account-service/service/internal/callers"
	"account-service/service/internal/config"
	"account-service/service/internal/leases"
	"account-service/service/internal/logging"
	"account-service/service/internal/security"

	"github.com/gofiber/fiber/v3"
)

func init() {
	os.Setenv("PROXY_API_KEY", "local-test-proxy-key")
}

func main() {
	cfg, err := config.Load()
	logger := logging.New(os.Stdout, "info")
	if err != nil {
		logger.Fatal().Err(err).Msg("load config")
	}
	logger = logging.New(os.Stdout, cfg.LogLevel)

	fiberApp, err := buildApp(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("build app")
	}

	addr := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	if err := fiberApp.Listen(addr); err != nil {
		logger.Fatal().Err(err).Str("addr", addr).Msg("listen")
	}
}

func buildApp(cfg config.Config) (*fiber.App, error) {
	codec, err := security.NewCredentialCodec(cfg.SecretEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("create credential codec: %w", err)
	}

	auditWriter := audit.NewMemoryWriter()
	accountService := accounts.NewService(accounts.NewMemoryRepository(codec), codec, auditWriter)
	leaseService := leases.NewService(accountService, cfg.DefaultLeaseTTL, cfg.MaxLeaseTTL, auditWriter)
	callerStore := callers.NewMemoryStore()

	adminStore := admin.NewMemoryStore(12 * time.Hour)
	passwordHash, err := security.HashPassword("strongpass")
	if err != nil {
		return nil, fmt.Errorf("hash default admin password: %w", err)
	}
	adminStore.AddUser(admin.User{
		ID:           "admin",
		Username:     "admin",
		PasswordHash: passwordHash,
		Status:       "active",
	})
	adminService := admin.NewService(adminStore, cfg.AdminSessionSecret, cfg.JWTAccessTokenTTL, cfg.JWTRefreshTokenTTL)

	return app.New(app.Options{
		HealthChecker:  healthCheckerFunc(func(context.Context) error { return nil }),
		AdminService:   adminService,
		AccountService: accountService,
		LeaseService:   leaseService,
		CallerStore:    callerStore,
		CORSOrigins:    cfg.CORSAllowedOrigins,
	}), nil
}

type healthCheckerFunc func(context.Context) error

func (fn healthCheckerFunc) Check(ctx context.Context) error {
	return fn(ctx)
}
