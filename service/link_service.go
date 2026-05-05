package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"

	"project/storage"
)

var (
	ErrInvalidURL = errors.New("invalid url")
)

type LinkService struct {
	storage    storage.Storage
	sendEveryN int
	log        *slog.Logger
}

type SaveResult struct {
	Count      int
	Duplicate  bool
	RandomLink string
}

func NewLinkService(storage storage.Storage, sendEveryN int, log *slog.Logger) (*LinkService, error) {
	if storage == nil {
		return nil, errors.New("storage is required")
	}

	if sendEveryN <= 0 {
		return nil, fmt.Errorf("sendEveryN must be greater than zero")
	}

	if log == nil {
		log = slog.Default()
	}

	return &LinkService{
		storage:    storage,
		sendEveryN: sendEveryN,
		log:        log,
	}, nil
}

func (s *LinkService) SaveLink(ctx context.Context, userID int64, rawLink string) (SaveResult, error) {
	link, err := NormalizeURL(rawLink)
	if err != nil {
		return SaveResult{}, err
	}

	if err := s.storage.SaveLink(ctx, userID, link); err != nil {
		if errors.Is(err, storage.ErrDuplicateLink) {
			s.log.InfoContext(ctx, "duplicate link ignored", "user_id", userID)
			return SaveResult{Duplicate: true}, nil
		}

		return SaveResult{}, fmt.Errorf("save link: %w", err)
	}

	s.log.InfoContext(ctx, "link saved", "user_id", userID)

	count, err := s.storage.IncrementUserMessages(ctx, userID)
	if err != nil {
		return SaveResult{}, fmt.Errorf("increment user messages: %w", err)
	}

	result := SaveResult{Count: count}
	if count%s.sendEveryN != 0 {
		return result, nil
	}

	randomLink, err := s.storage.GetRandomLink(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrNoSavedPages) {
			s.log.WarnContext(ctx, "no links available for scheduled random send", "user_id", userID)
			return result, nil
		}

		return SaveResult{}, fmt.Errorf("get random link: %w", err)
	}

	result.RandomLink = randomLink
	s.log.InfoContext(ctx, "scheduled random link selected", "user_id", userID, "count", count)
	return result, nil
}

func (s *LinkService) RandomLink(ctx context.Context, userID int64) (string, error) {
	link, err := s.storage.GetRandomLink(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get random link: %w", err)
	}

	return link, nil
}

func NormalizeURL(rawLink string) (string, error) {
	rawLink = strings.TrimSpace(rawLink)
	if rawLink == "" {
		return "", ErrInvalidURL
	}

	u, err := url.Parse(rawLink)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	u.Scheme = strings.ToLower(u.Scheme)
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("%w: only http and https URLs are supported", ErrInvalidURL)
	}

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return "", fmt.Errorf("%w: host is required", ErrInvalidURL)
	}

	u.Host = normalizedHost(u.Scheme, host, u.Port())
	if u.Path == "/" {
		u.Path = ""
		u.RawPath = ""
	}
	u.Fragment = ""
	return u.String(), nil
}

func normalizedHost(scheme string, host string, port string) string {
	if port == "" {
		return host
	}

	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		return host
	}

	return net.JoinHostPort(host, port)
}
