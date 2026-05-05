package telegram

import (
	"context"
	"testing"

	tgclient "project/clients/telegram"
	"project/events"
)

func TestFetchSkipsUnsupportedUpdatesAndAdvancesOffset(t *testing.T) {
	client := &stubTelegramClient{
		responses: [][]tgclient.Update{
			{
				{ID: 1},
				{
					ID: 2,
					Message: &tgclient.IncomingMessage{
						Text: "https://example.com",
						From: tgclient.From{ID: 42, Username: "alice"},
						Chat: tgclient.Chat{ID: 100},
					},
				},
			},
			nil,
		},
	}

	p := New(client, nil, nil)

	got, err := p.Fetch(context.Background(), 10)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("len(Fetch()) = %d, want 1", len(got))
	}

	if got[0].Type != events.Message {
		t.Fatalf("event type = %v, want %v", got[0].Type, events.Message)
	}

	meta, ok := got[0].Meta.(Meta)
	if !ok {
		t.Fatalf("event meta type = %T, want %T", got[0].Meta, Meta{})
	}

	if meta.UserID != 42 || meta.ChatID != 100 || meta.Username != "alice" {
		t.Fatalf("meta = %+v, want user_id=42 chat_id=100 username=alice", meta)
	}

	if client.offsets[0] != 0 {
		t.Fatalf("first fetch offset = %d, want 0", client.offsets[0])
	}

	if _, err := p.Fetch(context.Background(), 10); err != nil {
		t.Fatalf("second Fetch() error = %v", err)
	}

	if client.offsets[1] != 3 {
		t.Fatalf("second fetch offset = %d, want 3", client.offsets[1])
	}
}

type stubTelegramClient struct {
	responses [][]tgclient.Update
	offsets   []int
}

func (s *stubTelegramClient) Updates(ctx context.Context, offset int, limit int) ([]tgclient.Update, error) {
	s.offsets = append(s.offsets, offset)
	if len(s.responses) == 0 {
		return nil, nil
	}

	res := s.responses[0]
	s.responses = s.responses[1:]
	return res, nil
}

func (s *stubTelegramClient) SendMessage(ctx context.Context, chatID int, text string) error {
	return nil
}
