package sqlite

import (
	"context"
	"errors"
	"testing"

	"project/storage"
)

func TestStorageSaveRandomAndCount(t *testing.T) {
	ctx := context.Background()
	s := newTestStorage(t)

	if err := s.SaveLink(ctx, 100, "https://example.com/1"); err != nil {
		t.Fatalf("save link: %v", err)
	}

	if err := s.SaveLink(ctx, 100, "https://example.com/2"); err != nil {
		t.Fatalf("save second link: %v", err)
	}

	link, err := s.GetRandomLink(ctx, 100)
	if err != nil {
		t.Fatalf("get random link: %v", err)
	}

	if link == "" {
		t.Fatal("random link is empty")
	}

	count, err := s.IncrementUserMessages(ctx, 100)
	if err != nil {
		t.Fatalf("increment messages: %v", err)
	}

	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	count, err = s.CountUserMessages(ctx, 100)
	if err != nil {
		t.Fatalf("count messages: %v", err)
	}

	if count != 1 {
		t.Fatalf("stored count = %d, want 1", count)
	}
}

func TestStorageDuplicateLink(t *testing.T) {
	ctx := context.Background()
	s := newTestStorage(t)

	if err := s.SaveLink(ctx, 100, "https://example.com/1"); err != nil {
		t.Fatalf("save link: %v", err)
	}

	err := s.SaveLink(ctx, 100, "https://example.com/1")
	if !errors.Is(err, storage.ErrDuplicateLink) {
		t.Fatalf("expected ErrDuplicateLink, got %v", err)
	}
}

func TestStorageEmptyUser(t *testing.T) {
	ctx := context.Background()
	s := newTestStorage(t)

	_, err := s.GetRandomLink(ctx, 404)
	if !errors.Is(err, storage.ErrNoSavedPages) {
		t.Fatalf("expected ErrNoSavedPages, got %v", err)
	}

	count, err := s.CountUserMessages(ctx, 404)
	if err != nil {
		t.Fatalf("count messages: %v", err)
	}

	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

func newTestStorage(t *testing.T) *Storage {
	t.Helper()

	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}

	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Fatalf("close storage: %v", err)
		}
	})

	if err := s.Init(context.Background()); err != nil {
		t.Fatalf("init storage: %v", err)
	}

	return s
}
