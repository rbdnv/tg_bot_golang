package event_consumer

import (
	"context"
	"log/slog"
	"project/events"
	"time"
)

type Consumer struct {
	fetcher   events.Fetcher
	processor events.Processor
	batchSize int
	log       *slog.Logger
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int, log *slog.Logger) Consumer {
	if log == nil {
		log = slog.Default()
	}

	return Consumer{
		fetcher:   fetcher,
		processor: processor,
		batchSize: batchSize,
		log:       log,
	}
}

func (c Consumer) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.log.InfoContext(ctx, "consumer stopped")
			return ctx.Err()
		default:
		}

		gotEvents, err := c.fetcher.Fetch(ctx, c.batchSize)
		if err != nil {
			c.log.ErrorContext(ctx, "consumer fetch failed", "error", err)
			waitForNextTick(ctx, ticker)
			continue
		}

		if len(gotEvents) == 0 {
			waitForNextTick(ctx, ticker)
			continue
		}

		c.handleEvents(ctx, gotEvents)
	}
}

func (c *Consumer) handleEvents(ctx context.Context, events []events.Event) {
	for _, event := range events {
		c.log.DebugContext(ctx, "handling event", "type", event.Type)

		if err := c.processor.Process(ctx, event); err != nil {
			c.log.ErrorContext(ctx, "handle event failed", "error", err)
			continue
		}
	}
}

func waitForNextTick(ctx context.Context, ticker *time.Ticker) {
	select {
	case <-ctx.Done():
	case <-ticker.C:
	}
}
