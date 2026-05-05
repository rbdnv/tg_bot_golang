package config

import "testing"

func TestLoadUsesTelegramHostFromEnv(t *testing.T) {
	t.Setenv("BOT_TOKEN", "token")
	t.Setenv("DATABASE_PATH", "data/sqlite/storage.db")
	t.Setenv("SEND_EVERY_N", "5")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("ENV", "dev")
	t.Setenv("TELEGRAM_HOST", "http://localhost:8081/api")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.TelegramHost != "http://localhost:8081/api" {
		t.Fatalf("telegram host = %q, want %q", cfg.TelegramHost, "http://localhost:8081/api")
	}
}
