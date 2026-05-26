package logging

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	defaultLoggerMu sync.RWMutex
	defaultLogger   = New(os.Stdout, "info")
)

func New(out io.Writer, level string) zerolog.Logger {
	parsed, err := zerolog.ParseLevel(strings.ToLower(strings.TrimSpace(level)))
	if err != nil {
		parsed = zerolog.InfoLevel
	}

	return zerolog.New(out).Level(parsed).With().Timestamp().Logger()
}

func SetDefault(logger zerolog.Logger) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	defaultLogger = logger
}

func Default() zerolog.Logger {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	return defaultLogger
}

func NewOutput(appEnv string, logDir string, now time.Time) (*os.File, string, error) {
	if strings.EqualFold(strings.TrimSpace(appEnv), "development") {
		return os.Stdout, "", nil
	}

	dir := strings.TrimSpace(logDir)
	if dir == "" {
		dir = "logs"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, "", err
	}

	path := filepath.Join(dir, now.Format("2006-01-02")+".log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, "", err
	}
	return file, path, nil
}
