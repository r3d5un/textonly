package utils

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
)

func ReadJSON(r io.ReadCloser, data interface{}) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(data)
	if err != nil {
		return err
	}

	return nil
}

type ContextKey string

const LoggerKey ContextKey = "logger"

// Embeds a logger in the given context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// LoggerFromContext attempts to extract an embedded logger from the
// given context. If no context is found, it returns the default logger
// registered for the application.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(LoggerKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return logger
}
