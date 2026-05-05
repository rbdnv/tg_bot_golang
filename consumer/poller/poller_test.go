package poller

import (
	"context"
	"errors"
	"testing"

	"project/events"
)

func TestPollerStopsBatchOnFirstProcessError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetcher := &stubFetcher{
		events: []events.Event{
			{Type: events.Message, Text: "first"},
			{Type: events.Message, Text: "second"},
		},
	}
	processor := &stubProcessor{
		processFn: func(event events.Event) error {
			cancel()
			return errors.New("boom")
		},
	}

	err := New(fetcher, processor, 10, nil).Start(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Start() error = %v, want context.Canceled", err)
	}

	if len(processor.processed) != 1 {
		t.Fatalf("processed %d events, want 1", len(processor.processed))
	}

	if processor.processed[0].Text != "first" {
		t.Fatalf("processed first event = %q, want %q", processor.processed[0].Text, "first")
	}
}

type stubFetcher struct {
	events []events.Event
	calls  int
}

func (s *stubFetcher) Fetch(ctx context.Context, limit int) ([]events.Event, error) {
	s.calls++
	if s.calls == 1 {
		return s.events, nil
	}

	<-ctx.Done()
	return nil, ctx.Err()
}

type stubProcessor struct {
	processed []events.Event
	processFn func(event events.Event) error
}

func (s *stubProcessor) Process(ctx context.Context, event events.Event) error {
	s.processed = append(s.processed, event)
	if s.processFn != nil {
		return s.processFn(event)
	}

	return nil
}
