package telegram

import (
	"context"
	"errors"
	"strings"

	"project/lib/e"
	"project/service"
	"project/storage"
)

const (
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start"
)

func (p *Processor) doCmd(ctx context.Context, text string, meta Meta) error {
	text = strings.TrimSpace(text)

	switch text {
	case RndCmd:
		return p.sendRandom(ctx, meta.ChatID, meta.UserID)
	case HelpCmd:
		return p.sendHelp(ctx, meta.ChatID)
	case StartCmd:
		return p.sendHello(ctx, meta.ChatID)
	}

	if strings.HasPrefix(text, "/") {
		return p.tg.SendMessage(ctx, meta.ChatID, msgUnknownCommand)
	}

	return p.savePage(ctx, meta.ChatID, meta.UserID, text)
}

func (p *Processor) savePage(ctx context.Context, chatID int, userID int64, pageURL string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: save page", err) }()

	result, err := p.service.SaveLink(ctx, userID, pageURL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			return p.tg.SendMessage(ctx, chatID, msgInvalidURL)
		}

		return err
	}

	if result.Duplicate {
		return p.tg.SendMessage(ctx, chatID, msgAlreadyExists)
	}

	if err := p.tg.SendMessage(ctx, chatID, msgSaved); err != nil {
		return err
	}

	if result.RandomLink == "" {
		return nil
	}

	return p.tg.SendMessage(ctx, chatID, result.RandomLink)
}

func (p *Processor) sendRandom(ctx context.Context, chatID int, userID int64) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: send random", err) }()

	link, err := p.service.RandomLink(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrNoSavedPages) {
			return p.tg.SendMessage(ctx, chatID, msgNoSavedPages)
		}

		return err
	}

	return p.tg.SendMessage(ctx, chatID, link)
}

func (p *Processor) sendHelp(ctx context.Context, chatID int) error {
	return p.tg.SendMessage(ctx, chatID, msgHelp)
}

func (p *Processor) sendHello(ctx context.Context, chatID int) error {
	return p.tg.SendMessage(ctx, chatID, msgHello)
}
