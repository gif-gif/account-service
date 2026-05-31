package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv                            string
	DatabaseURL                       string
	ServiceBaseURL                    string
	SecretEncryptionKey               string
	DefaultLeaseTTL                   time.Duration
	MaxLeaseTTL                       time.Duration
	LeaseCleanupInterval              time.Duration
	AdminSessionSecret                string
	JWTAccessTokenTTL                 time.Duration
	JWTRefreshTokenTTL                time.Duration
	CORSAllowedOrigins                []string
	LogLevel                          string
	LogDir                            string
	KiroLoginFeishuWebhook            string
	ExternalAPIKeyAuthEnabled         bool
	HealthCheckDatabaseTimeout        time.Duration
	HTTPHost                          string
	HTTPPort                          int
	DefaultLeaseTTLSeconds            int
	MaxLeaseTTLSeconds                int
	LeaseCleanupIntervalSeconds       int
	JWTAccessTokenTTLSeconds          int
	JWTRefreshTokenTTLSeconds         int
	HealthCheckDatabaseTimeoutSeconds int
}

func Load() (Config, error) {
	loadEnvFiles()

	cfg := Config{
		AppEnv:                            envString("APP_ENV", "local"),
		DatabaseURL:                       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		ServiceBaseURL:                    strings.TrimSpace(os.Getenv("SERVICE_BASE_URL")),
		SecretEncryptionKey:               strings.TrimSpace(os.Getenv("SECRET_ENCRYPTION_KEY")),
		AdminSessionSecret:                strings.TrimSpace(os.Getenv("ADMIN_SESSION_SECRET")),
		CORSAllowedOrigins:                splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
		LogLevel:                          envString("LOG_LEVEL", "info"),
		LogDir:                            envString("LOG_DIR", "logs"),
		KiroLoginFeishuWebhook:            strings.TrimSpace(os.Getenv("KIRO_LOGIN_FEISHU_WEBHOOK")),
		HTTPHost:                          envString("HTTP_HOST", "127.0.0.1"),
		HTTPPort:                          envInt("HTTP_PORT", 8000),
		DefaultLeaseTTLSeconds:            envInt("DEFAULT_LEASE_TTL_SECONDS", 900),
		MaxLeaseTTLSeconds:                envInt("MAX_LEASE_TTL_SECONDS", 7200),
		LeaseCleanupIntervalSeconds:       envInt("LEASE_CLEANUP_INTERVAL_SECONDS", 60),
		JWTAccessTokenTTLSeconds:          envInt("JWT_ACCESS_TOKEN_TTL_SECONDS", 172800),
		JWTRefreshTokenTTLSeconds:         envInt("JWT_REFRESH_TOKEN_TTL_SECONDS", 604800),
		HealthCheckDatabaseTimeoutSeconds: envInt("HEALTH_CHECK_DATABASE_TIMEOUT_SECONDS", 3),
	}
	cfg.ExternalAPIKeyAuthEnabled = envBool("EXTERNAL_API_KEY_AUTH_ENABLED", cfg.AppEnv != "local")

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.SecretEncryptionKey == "" {
		return Config{}, errors.New("SECRET_ENCRYPTION_KEY is required")
	}
	if cfg.AdminSessionSecret == "" {
		return Config{}, errors.New("ADMIN_SESSION_SECRET is required")
	}
	if cfg.DefaultLeaseTTLSeconds <= 0 {
		return Config{}, errors.New("DEFAULT_LEASE_TTL_SECONDS must be greater than 0")
	}
	if cfg.MaxLeaseTTLSeconds <= 0 {
		return Config{}, errors.New("MAX_LEASE_TTL_SECONDS must be greater than 0")
	}
	if cfg.LeaseCleanupIntervalSeconds <= 0 {
		return Config{}, errors.New("LEASE_CLEANUP_INTERVAL_SECONDS must be greater than 0")
	}
	if cfg.JWTAccessTokenTTLSeconds <= 0 {
		return Config{}, errors.New("JWT_ACCESS_TOKEN_TTL_SECONDS must be greater than 0")
	}
	if cfg.JWTRefreshTokenTTLSeconds <= 0 {
		return Config{}, errors.New("JWT_REFRESH_TOKEN_TTL_SECONDS must be greater than 0")
	}
	if cfg.HealthCheckDatabaseTimeoutSeconds <= 0 {
		return Config{}, errors.New("HEALTH_CHECK_DATABASE_TIMEOUT_SECONDS must be greater than 0")
	}
	if cfg.DefaultLeaseTTLSeconds > cfg.MaxLeaseTTLSeconds {
		return Config{}, fmt.Errorf("DEFAULT_LEASE_TTL_SECONDS must be less than or equal to MAX_LEASE_TTL_SECONDS")
	}

	cfg.DefaultLeaseTTL = time.Duration(cfg.DefaultLeaseTTLSeconds) * time.Second
	cfg.MaxLeaseTTL = time.Duration(cfg.MaxLeaseTTLSeconds) * time.Second
	cfg.LeaseCleanupInterval = time.Duration(cfg.LeaseCleanupIntervalSeconds) * time.Second
	cfg.JWTAccessTokenTTL = time.Duration(cfg.JWTAccessTokenTTLSeconds) * time.Second
	cfg.JWTRefreshTokenTTL = time.Duration(cfg.JWTRefreshTokenTTLSeconds) * time.Second
	cfg.HealthCheckDatabaseTimeout = time.Duration(cfg.HealthCheckDatabaseTimeoutSeconds) * time.Second

	return cfg, nil
}

func loadEnvFiles() {
	for _, path := range candidateEnvFiles() {
		loadEnvFile(path)
	}
}

func candidateEnvFiles() []string {
	appEnv := strings.TrimSpace(os.Getenv("APP_ENV"))
	if appEnv == "" {
		appEnv = "local"
	}

	names := make([]string, 0, 4)
	for _, key := range []string{"SERVICE_ENV_FILE", "ENV_FILE"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			names = append(names, value)
		}
	}
	names = append(names, ".env."+appEnv, ".env.local", ".env")

	seen := map[string]bool{}
	out := make([]string, 0, len(names)*3)
	for _, name := range names {
		for _, path := range candidateEnvPaths(name) {
			cleaned := filepath.Clean(path)
			if !seen[cleaned] {
				seen[cleaned] = true
				out = append(out, cleaned)
			}
		}
	}
	return out
}

func candidateEnvPaths(name string) []string {
	if filepath.IsAbs(name) {
		return []string{name}
	}
	return []string{
		name,
		filepath.Join("..", name),
		filepath.Join("/app", name),
	}
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, ok := parseEnvLine(scanner.Text())
		if !ok {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

func parseEnvLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	line = strings.TrimPrefix(line, "export ")
	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return "", "", false
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", "", false
	}
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		first := value[0]
		last := value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			value = value[1 : len(value)-1]
		}
	}
	return key, value, true
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
