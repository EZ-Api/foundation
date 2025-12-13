package logging

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Options struct {
	Service string
}

func New(opts Options) (*slog.Logger, zerolog.Logger) {
	level := parseLevel(strings.TrimSpace(os.Getenv("EZ_LOG_LEVEL")))
	zerolog.SetGlobalLevel(toZerologLevel(level))

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	zl := zerolog.New(output).
		Level(toZerologLevel(level)).
		With().
		Timestamp().
		Str("service", strings.TrimSpace(opts.Service)).
		Logger()

	sl := slog.New(NewZerologHandler(zl, level))
	slog.SetDefault(sl)
	return sl, zl
}

func parseLevel(raw string) slog.Level {
	switch strings.ToLower(raw) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info", "":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

func toZerologLevel(level slog.Level) zerolog.Level {
	switch {
	case level <= slog.LevelDebug:
		return zerolog.DebugLevel
	case level <= slog.LevelInfo:
		return zerolog.InfoLevel
	case level <= slog.LevelWarn:
		return zerolog.WarnLevel
	default:
		return zerolog.ErrorLevel
	}
}

