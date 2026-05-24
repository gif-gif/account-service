package logging

import (
	"io"
	"strings"

	"github.com/rs/zerolog"
)

func New(out io.Writer, level string) zerolog.Logger {
	parsed, err := zerolog.ParseLevel(strings.ToLower(strings.TrimSpace(level)))
	if err != nil {
		parsed = zerolog.InfoLevel
	}

	return zerolog.New(out).Level(parsed).With().Timestamp().Logger()
}
