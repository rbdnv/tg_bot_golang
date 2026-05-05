package poller

import (
	"context"
	"log/slog"
	"time"

	"project/events"
)

type Poller struct {
	fetcher   events.Fetcher
	processor events.Processor
	batchSize int
	log       *slog.Logger
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int, log *slog.Logger) *Poller {
	if log == nil {
		log = slog.Default()
	}

	return &Poller{
		fetcher:   fetcher,
		processor: processor,
		batchSize: batchSize,
		log:       log,
	}
}

func (p *Poller) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.log.InfoContext(ctx, "poller stopped")
			return ctx.Err()
		default:
		}

		gotEvents, err := p.fetcher.Fetch(ctx, p.batchSize)
		if err != nil {
			p.log.ErrorContext(ctx, "poller fetch failed", "error", err)
			waitForNextTick(ctx, ticker)
			continue
		}

		if len(gotEvents) == 0 {
			waitForNextTick(ctx, ticker)
			continue
		}

		p.handleEvents(ctx, gotEvents)
	}
}

func (p *Poller) handleEvents(ctx context.Context, events []events.Event) {
	for _, event := range events {
		p.log.DebugContext(ctx, "handling event", "type", event.Type)

		if err := p.processor.Process(ctx, event); err != nil {
			p.log.ErrorContext(ctx, "handle event failed", "error", err)
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
