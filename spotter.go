package spotter

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

// Provider defines the target environment to adjust log formatting
type Provider string

const (
	ProviderGCP   Provider = "gcp"
	ProviderAzure Provider = "azure"
	ProviderLocal Provider = "local"
)

// Config holds all necessary setup. No hidden ENV reads here.
type Config struct {
	Provider  Provider
	Level     slog.Level
	AddSource bool
	// StaticAttrs allows you to inject constant metadata (e.g., App Name, Revision)
	// This replaces the hardcoded calls to os.Getenv for metadata
	StaticAttrs map[string]string
}

// New creates a configured logger instance
func New(cfg Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 1. Universal Time Handling (UTC is King)
			if a.Key == slog.TimeKey {
				t := a.Value.Time().UTC()
				return slog.String(slog.TimeKey, t.Format(time.RFC3339Nano))
			}

			// 2. Provider Specific Logic
			switch cfg.Provider {
			case ProviderGCP:
				return gcpTransformer(a)
			case ProviderAzure:
				return azureTransformer(a)
			}

			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	// Create the base logger
	logger := slog.New(handler)

	// Inject the static attributes (metadata) if any exist
	if len(cfg.StaticAttrs) > 0 {
		attrs := make([]any, 0, len(cfg.StaticAttrs))
		for k, v := range cfg.StaticAttrs {
			attrs = append(attrs, slog.String(k, v))
		}
		logger = logger.With(attrs...)
	}

	// Set as global default
	slog.SetDefault(logger)

	return logger
}

// gcpTransformer handles Google Cloud Run specific keys
func gcpTransformer(a slog.Attr) slog.Attr {
	// Rename "msg" -> "message"
	if a.Key == slog.MessageKey {
		return slog.Attr{Key: "message", Value: a.Value}
	}

	// Rename "source" -> "logging.googleapis.com/sourceLocation"
	if a.Key == slog.SourceKey {
		return slog.Attr{Key: "logging.googleapis.com/sourceLocation", Value: a.Value}
	}

	// Map "level" -> "severity" (Uppercase)
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		var severity string
		switch {
		case level < slog.LevelInfo:
			severity = "DEBUG"
		case level < slog.LevelWarn:
			severity = "INFO"
		case level < slog.LevelError:
			severity = "WARNING"
		default:
			severity = "ERROR"
		}
		return slog.Attr{Key: "severity", Value: slog.StringValue(severity)}
	}
	return a
}

// azureTransformer handles Azure Monitor/ACA specific preferences
func azureTransformer(a slog.Attr) slog.Attr {
	// Azure often prefers lowercase levels for query consistency
	if a.Key == slog.LevelKey {
		return slog.String(slog.LevelKey, strings.ToLower(a.Value.String()))
	}
	return a
}
