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
	"account-service/service/internal/db"
	"account-service/service/internal/leases"
	"account-service/service/internal/logging"
	"account-service/service/internal/modelconfig"
	"account-service/service/internal/security"

	"github.com/gofiber/fiber/v3"
)

//	func init() {
//		os.Setenv("APP_ENV", "local")
//	}
func main() {
	defer func() {
		fmt.Print("Shutting down1...", time.Now().String())
	}()

	cfg, err := config.Load()
	logger := logging.New(os.Stdout, "info")
	if err != nil {
		logger.Fatal().Err(err).Msg("load config")
	}
	logOutput, logPath, err := logging.NewOutput(cfg.AppEnv, cfg.LogDir, time.Now())
	if err != nil {
		logger.Fatal().Err(err).Msg("open log output")
	}
	if logOutput != os.Stdout {
		defer logOutput.Close()
	}
	logger = logging.New(logOutput, cfg.LogLevel)
	logging.SetDefault(logger)
	if logPath != "" {
		logger.Info().Str("path", logPath).Msg("service logs writing to file")
	}

	fiberApp, err := buildApp(cfg)
	if err != nil {
		fmt.Print("build app error: ", err.Error())
		logger.Fatal().Err(err).Msg("build app")
	}

	addr := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	fmt.Printf("Listening on %s, %s\n", addr, time.Now().String())
	if err := fiberApp.Listen(addr); err != nil {
		fmt.Print("Shutting down...", time.Now().String())
		logger.Fatal().Err(err).Str("addr", addr).Msg("listen")
		return
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
	var callerStore callers.Store = callers.NewMemoryStore()
	modelConfigService := modelconfig.NewService(modelconfig.NewMemoryRepository(nil))
	if cfg.DatabaseURL != "" {
		pool, err := db.Open(context.Background(), cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		if err := db.ApplyMigrations(context.Background(), pool); err != nil {
			pool.Close()
			return nil, err
		}
		callerStore = callers.NewPostgresStore(pool)
		modelConfigService = modelconfig.NewService(modelconfig.NewPostgresRepository(pool))
	}

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
		HealthChecker:             healthCheckerFunc(func(context.Context) error { return nil }),
		AdminService:              adminService,
		AccountService:            accountService,
		LeaseService:              leaseService,
		CallerStore:               callerStore,
		ModelConfig:               modelConfigService,
		ExternalAPIKeyAuthEnabled: &cfg.ExternalAPIKeyAuthEnabled,
		CORSOrigins:               cfg.CORSAllowedOrigins,
	}), nil
}

type healthCheckerFunc func(context.Context) error

func (fn healthCheckerFunc) Check(ctx context.Context) error {
	return fn(ctx)
}
