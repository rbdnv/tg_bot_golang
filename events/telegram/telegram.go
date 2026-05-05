package telegram

import (
	"context"
	"errors"
	"log/slog"

	tgclient "project/clients/telegram"
	"project/events"
	"project/lib/e"
	"project/service"
)

type Processor struct {
	tg          telegramClient
	offsetStore offsetStore
	offset      int
	service     *service.LinkService
	log         *slog.Logger
}

type telegramClient interface {
	Updates(ctx context.Context, offset int, limit int) ([]tgclient.Update, error)
	SendMessage(ctx context.Context, chatID int, text string) error
}

type offsetStore interface {
	SaveTelegramOffset(ctx context.Context, offset int) error
}

type Meta struct {
	UpdateID int
	ChatID   int
	UserID   int64
	Username string
}

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
)

func New(client telegramClient, service *service.LinkService, offsetStore offsetStore, initialOffset int, log *slog.Logger) *Processor {
	if log == nil {
		log = slog.Default()
	}

	return &Processor{
		tg:          client,
		offsetStore: offsetStore,
		offset:      initialOffset,
		service:     service,
		log:         log,
	}
}

func (p *Processor) Fetch(ctx context.Context, limit int) ([]events.Event, error) {
	updates, err := p.tg.Updates(ctx, p.offset, limit)
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))
	skipped := 0

	for _, u := range updates {
		if event, ok := event(u); ok {
			res = append(res, event)
			continue
		}

		skipped++
	}

	if skipped > 0 {
		p.log.DebugContext(ctx, "skipped unsupported telegram updates", "count", skipped)
	}

	return res, nil
}

func (p *Processor) Process(ctx context.Context, event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(ctx, event)
	default:
		return e.Wrap("can't process message", ErrUnknownEventType)
	}
}

func (p *Processor) processMessage(ctx context.Context, event events.Event) error {
	meta, err := meta(event)

	if err != nil {
		return e.Wrap("can't process message", err)
	}

	p.log.InfoContext(ctx, "incoming telegram message", "chat_id", meta.ChatID, "user_id", meta.UserID, "username", meta.Username)

	if err := p.doCmd(ctx, event.Text, meta); err != nil {
		return e.Wrap("can't process message", err)
	}

	if err := p.confirmUpdate(ctx, meta.UpdateID+1); err != nil {
		return e.Wrap("can't process message", err)
	}

	return nil
}

func (p *Processor) confirmUpdate(ctx context.Context, offset int) error {
	if p.offsetStore == nil {
		p.offset = offset
		return nil
	}

	if err := p.offsetStore.SaveTelegramOffset(ctx, offset); err != nil {
		return e.Wrap("can't save telegram offset", err)
	}

	p.offset = offset
	return nil
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknownMetaType)
	}

	return res, nil
}

func event(upd tgclient.Update) (events.Event, bool) {
	if upd.Message == nil {
		return events.Event{}, false
	}

	return events.Event{
		Type: events.Message,
		Text: upd.Message.Text,
		Meta: Meta{
			UpdateID: upd.ID,
			ChatID:   upd.Message.Chat.ID,
			UserID:   upd.Message.From.ID,
			Username: upd.Message.From.Username,
		},
	}, true
}
