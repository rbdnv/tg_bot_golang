package files

import (
	"context"
	"errors"
	"testing"

	"project/storage"
)

func TestStorageSaveRandomAndCount(t *testing.T) {
	ctx := context.Background()
	s := New(t.TempDir())

	if err := s.Init(ctx); err != nil {
		t.Fatalf("init storage: %v", err)
	}

	if err := s.SaveLink(ctx, 7, "https://example.com/1"); err != nil {
		t.Fatalf("save first link: %v", err)
	}

	if err := s.SaveLink(ctx, 7, "https://example.com/2"); err != nil {
		t.Fatalf("save second link: %v", err)
	}

	link, err := s.GetRandomLink(ctx, 7)
	if err != nil {
		t.Fatalf("get random link: %v", err)
	}

	if link != "https://example.com/1" && link != "https://example.com/2" {
		t.Fatalf("random link = %q, want one of saved links", link)
	}

	count, err := s.IncrementUserMessages(ctx, 7)
	if err != nil {
		t.Fatalf("increment messages: %v", err)
	}

	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	count, err = s.CountUserMessages(ctx, 7)
	if err != nil {
		t.Fatalf("count messages: %v", err)
	}

	if count != 1 {
		t.Fatalf("stored count = %d, want 1", count)
	}
}

func TestStorageDuplicateLink(t *testing.T) {
	ctx := context.Background()
	s := New(t.TempDir())

	if err := s.Init(ctx); err != nil {
		t.Fatalf("init storage: %v", err)
	}

	if err := s.SaveLink(ctx, 7, "https://example.com/1"); err != nil {
		t.Fatalf("save link: %v", err)
	}

	err := s.SaveLink(ctx, 7, "https://example.com/1")
	if !errors.Is(err, storage.ErrDuplicateLink) {
		t.Fatalf("expected ErrDuplicateLink, got %v", err)
	}
}

func TestStorageEmptyUser(t *testing.T) {
	ctx := context.Background()
	s := New(t.TempDir())

	if err := s.Init(ctx); err != nil {
		t.Fatalf("init storage: %v", err)
	}

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
