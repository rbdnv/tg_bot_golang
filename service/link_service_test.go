package service

import (
	"context"
	"errors"
	"testing"

	"project/storage"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "https url", raw: " https://example.com/path#fragment ", want: "https://example.com/path"},
		{name: "http url", raw: "http://example.com", want: "http://example.com"},
		{name: "missing scheme", raw: "example.com", wantErr: true},
		{name: "unsupported scheme", raw: "ftp://example.com", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeURL(tt.raw)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidURL) {
					t.Fatalf("expected ErrInvalidURL, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSaveLinkSendsRandomEveryN(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStorage()
	svc, err := NewLinkService(store, 2, nil)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	first, err := svc.SaveLink(ctx, 42, "https://example.com/1")
	if err != nil {
		t.Fatalf("first save: %v", err)
	}

	if first.Count != 1 {
		t.Fatalf("first count = %d, want 1", first.Count)
	}

	if first.RandomLink != "" {
		t.Fatalf("first random link = %q, want empty", first.RandomLink)
	}

	second, err := svc.SaveLink(ctx, 42, "https://example.com/2")
	if err != nil {
		t.Fatalf("second save: %v", err)
	}

	if second.Count != 2 {
		t.Fatalf("second count = %d, want 2", second.Count)
	}

	if second.RandomLink == "" {
		t.Fatal("second random link is empty")
	}
}

func TestSaveLinkDuplicateIsIgnored(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStorage()
	svc, err := NewLinkService(store, 2, nil)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if _, err := svc.SaveLink(ctx, 42, "https://example.com/1"); err != nil {
		t.Fatalf("first save: %v", err)
	}

	result, err := svc.SaveLink(ctx, 42, "https://example.com/1")
	if err != nil {
		t.Fatalf("duplicate save: %v", err)
	}

	if !result.Duplicate {
		t.Fatal("duplicate flag is false")
	}

	count, err := svc.CountUserMessages(ctx, 42)
	if err != nil {
		t.Fatalf("count messages: %v", err)
	}

	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

type memoryStorage struct {
	links  map[int64]map[string]struct{}
	counts map[int64]int
}

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{
		links:  make(map[int64]map[string]struct{}),
		counts: make(map[int64]int),
	}
}

func (s *memoryStorage) Init(ctx context.Context) error {
	return nil
}

func (s *memoryStorage) SaveLink(ctx context.Context, userID int64, link string) error {
	if s.links[userID] == nil {
		s.links[userID] = make(map[string]struct{})
	}

	if _, ok := s.links[userID][link]; ok {
		return storage.ErrDuplicateLink
	}

	s.links[userID][link] = struct{}{}
	return nil
}

func (s *memoryStorage) GetRandomLink(ctx context.Context, userID int64) (string, error) {
	for link := range s.links[userID] {
		return link, nil
	}

	return "", storage.ErrNoSavedPages
}

func (s *memoryStorage) IncrementUserMessages(ctx context.Context, userID int64) (int, error) {
	s.counts[userID]++
	return s.counts[userID], nil
}

func (s *memoryStorage) CountUserMessages(ctx context.Context, userID int64) (int, error) {
	return s.counts[userID], nil
}

func (s *memoryStorage) Close() error {
	return nil
}

var _ storage.Storage = (*memoryStorage)(nil)
