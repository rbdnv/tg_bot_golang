package telegram

import (
	"context"
	"errors"
	"testing"

	tgclient "project/clients/telegram"
	"project/events"
)

func TestProcessConfirmsOffsetAfterSuccessfulCommand(t *testing.T) {
	ctx := context.Background()
	client := &stubTelegramClient{responses: [][]tgclient.Update{nil}}
	store := &stubOffsetStore{}
	p := New(client, nil, store, 2, nil)

	err := p.Process(ctx, events.Event{
		Type: events.Message,
		Text: HelpCmd,
		Meta: Meta{
			UpdateID: 2,
			ChatID:   99,
		},
	})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if len(store.offsets) != 1 || store.offsets[0] != 3 {
		t.Fatalf("saved offsets = %v, want [3]", store.offsets)
	}

	if _, err := p.Fetch(ctx, 10); err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if len(client.offsets) != 1 || client.offsets[0] != 3 {
		t.Fatalf("fetch offsets = %v, want [3]", client.offsets)
	}
}

func TestProcessDoesNotConfirmOffsetOnCommandFailure(t *testing.T) {
	client := &stubTelegramClient{sendErr: errors.New("send failed")}
	store := &stubOffsetStore{}
	p := New(client, nil, store, 7, nil)

	err := p.Process(context.Background(), events.Event{
		Type: events.Message,
		Text: HelpCmd,
		Meta: Meta{
			UpdateID: 7,
			ChatID:   99,
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if len(store.offsets) != 0 {
		t.Fatalf("saved offsets = %v, want none", store.offsets)
	}
}

type stubOffsetStore struct {
	offsets []int
}

func (s *stubOffsetStore) SaveTelegramOffset(ctx context.Context, offset int) error {
	s.offsets = append(s.offsets, offset)
	return nil
}
