package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	tgClient "project/clients/Telegram"
	"project/config"
	event_consumer "project/consumer/event-consumer"
	"project/events/telegram"
	"project/service"
	"project/storage/sqlite"
)

const batchSize = 100

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	log := newLogger(cfg)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	storage, err := sqlite.New(cfg.DatabasePath)
	if err != nil {
		log.Error("connect to storage failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			log.Error("close storage failed", "error", err)
		}
	}()

	if err := storage.Init(ctx); err != nil {
		log.Error("init storage failed", "error", err)
		os.Exit(1)
	}

	linkService, err := service.NewLinkService(storage, cfg.SendEveryN, log)
	if err != nil {
		log.Error("create service failed", "error", err)
		os.Exit(1)
	}

	eventsProcessor := telegram.New(
		tgClient.New(cfg.TelegramHost, cfg.BotToken),
		linkService,
		log,
	)

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize, log)
	log.Info("service started", "env", cfg.Env, "send_every_n", cfg.SendEveryN)

	if err := consumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("service stopped with error", "error", err)
		os.Exit(1)
	}

	log.Info("service stopped")
}

func newLogger(cfg config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	if cfg.Env == "local" || cfg.Env == "dev" {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
