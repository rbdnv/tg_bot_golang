package storage

import (
	"context"
	"errors"
)

type Storage interface {
	Init(ctx context.Context) error
	SaveLink(ctx context.Context, userID int64, link string) error
	GetRandomLink(ctx context.Context, userID int64) (string, error)
	IncrementUserMessages(ctx context.Context, userID int64) (int, error)
	Close() error
}

var (
	ErrNoSavedPages  = errors.New("no saved pages")
	ErrDuplicateLink = errors.New("duplicate link")
)
