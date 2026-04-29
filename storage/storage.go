package storage

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"project/lib/e"
	"time"
)

type Storage interface {
	Init(ctx context.Context) error
	SaveLink(ctx context.Context, userID int64, link string) error
	GetRandomLink(ctx context.Context, userID int64) (string, error)
	IncrementUserMessages(ctx context.Context, userID int64) (int, error)
	CountUserMessages(ctx context.Context, userID int64) (int, error)
	Close() error
}

var (
	ErrNoSavedPages  = errors.New("no saved pages")
	ErrDuplicateLink = errors.New("duplicate link")
)

type Page struct {
	UserID    int64
	URL       string
	UserName  string
	Link      string
	CreatedAt time.Time
}

func (p Page) Hash() (string, error) {
	h := sha1.New()

	if _, err := io.WriteString(h, firstNonEmpty(p.Link, p.URL)); err != nil {
		return "", e.Wrap("can't calculate hash", err)
	}

	if _, err := io.WriteString(h, fmt.Sprintf("%d:%s", p.UserID, p.UserName)); err != nil {
		return "", e.Wrap("can't calculate hash", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
