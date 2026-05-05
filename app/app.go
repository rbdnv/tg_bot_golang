package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	telegramclient "project/clients/telegram"
	"project/config"
	"project/consumer/poller"
	telegramevents "project/events/telegram"
	"project/service"
	"project/storage/sqlite"
)

const batchSize = 100

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := newLogger(cfg)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return run(ctx, cfg, log)
}

func run(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	store, err := sqlite.New(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("connect to storage: %w", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Error("close storage failed", "error", err)
		}
	}()

	if err := store.Init(ctx); err != nil {
		return fmt.Errorf("init storage: %w", err)
	}

	linkService, err := service.NewLinkService(store, cfg.SendEveryN, log)
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	processor := telegramevents.New(
		telegramclient.New(cfg.TelegramHost, cfg.BotToken),
		linkService,
		log,
	)

	log.Info("service started", "env", cfg.Env, "send_every_n", cfg.SendEveryN)

	if err := poller.New(processor, processor, batchSize, log).Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("run poller: %w", err)
	}

	log.Info("service stopped")
	return nil
}

func newLogger(cfg config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	if cfg.Env == "local" || cfg.Env == "dev" {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
