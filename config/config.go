package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const defaultTelegramHost = "api.telegram.org"

type Config struct {
	BotToken     string
	DatabasePath string
	SendEveryN   int
	LogLevel     slog.Level
	Env          string
	TelegramHost string
}

func Load() (Config, error) {
	cfg := Config{
		BotToken:     strings.TrimSpace(os.Getenv("BOT_TOKEN")),
		DatabasePath: firstNonEmpty(os.Getenv("DATABASE_PATH"), os.Getenv("DATABASE_URL")),
		Env:          firstNonEmpty(os.Getenv("ENV"), "local"),
		TelegramHost: firstNonEmpty(os.Getenv("TELEGRAM_HOST"), defaultTelegramHost),
	}

	if cfg.BotToken == "" {
		return Config{}, errors.New("BOT_TOKEN is required")
	}

	if cfg.DatabasePath == "" {
		return Config{}, errors.New("DATABASE_PATH or DATABASE_URL is required")
	}

	sendEveryN, err := parsePositiveInt("SEND_EVERY_N", os.Getenv("SEND_EVERY_N"))
	if err != nil {
		return Config{}, err
	}
	cfg.SendEveryN = sendEveryN

	level, err := parseLogLevel(firstNonEmpty(os.Getenv("LOG_LEVEL"), "info"))
	if err != nil {
		return Config{}, err
	}
	cfg.LogLevel = level

	return cfg, nil
}

func parsePositiveInt(name string, value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("%s is required", name)
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}

	if n <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}

	return n, nil
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported LOG_LEVEL %q", value)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}
