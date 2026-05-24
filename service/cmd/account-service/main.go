package main

import (
	"context"
	"fmt"
	"os"

	"account-service/service/internal/app"
	"account-service/service/internal/config"
	"account-service/service/internal/logging"
)

func main() {
	cfg, err := config.Load()
	logger := logging.New(os.Stdout, "info")
	if err != nil {
		logger.Fatal().Err(err).Msg("load config")
	}
	logger = logging.New(os.Stdout, cfg.LogLevel)

	fiberApp := app.New(app.Options{
		HealthChecker: healthCheckerFunc(func(context.Context) error { return nil }),
	})

	addr := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	if err := fiberApp.Listen(addr); err != nil {
		logger.Fatal().Err(err).Str("addr", addr).Msg("listen")
	}
}

type healthCheckerFunc func(context.Context) error

func (fn healthCheckerFunc) Check(ctx context.Context) error {
	return fn(ctx)
}
