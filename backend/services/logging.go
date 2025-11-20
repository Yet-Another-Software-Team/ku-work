package services

import (
	"errors"
	"log/slog"
	"os"
)

func InitializeLoggingService() error {
	logger_type, has_logger_type := os.LookupEnv("LOGGER_TYPE")
	if !has_logger_type {
		return errors.New("no logger type specified")
	}
	switch logger_type {
	case "TEXT":
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
		return nil
	case "JSON":
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
		return nil
	default:
		return errors.New("unknown logger type")
	}
}
