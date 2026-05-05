package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		apiBase string
		token   string
		want    string
	}{
		{name: "host without scheme", apiBase: "api.telegram.org", token: "token", want: "https://api.telegram.org/bottoken"},
		{name: "full url with path", apiBase: "https://example.com/custom/api", token: "token", want: "https://example.com/custom/api/bottoken"},
		{name: "empty uses default", apiBase: "", token: "token", want: "https://api.telegram.org/bottoken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newBaseURL(tt.apiBase, tt.token)
			if got.String() != tt.want {
				t.Fatalf("newBaseURL() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestNewSetsDefaultTimeout(t *testing.T) {
	client := New("api.telegram.org", "token")
	if client.client.Timeout != defaultTimeout {
		t.Fatalf("timeout = %v, want %v", client.client.Timeout, defaultTimeout)
	}
}

func TestUpdatesUsesConfiguredBaseURL(t *testing.T) {
	var gotPath string
	var gotOffset string
	var gotLimit string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotOffset = r.URL.Query().Get("offset")
		gotLimit = r.URL.Query().Get("limit")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":[{"update_id":1,"message":{"text":"https://example.com","from":{"id":42,"username":"alice"},"chat":{"id":7}}}]}`))
	}))
	t.Cleanup(server.Close)

	client := New(server.URL+"/telegram", "test-token")

	updates, err := client.Updates(context.Background(), 123, 50)
	if err != nil {
		t.Fatalf("Updates() error = %v", err)
	}

	if gotPath != "/telegram/bottest-token/getUpdates" {
		t.Fatalf("path = %q, want %q", gotPath, "/telegram/bottest-token/getUpdates")
	}

	if gotOffset != "123" {
		t.Fatalf("offset = %q, want %q", gotOffset, "123")
	}

	if gotLimit != "50" {
		t.Fatalf("limit = %q, want %q", gotLimit, "50")
	}

	if len(updates) != 1 {
		t.Fatalf("len(Updates()) = %d, want 1", len(updates))
	}

	if updates[0].Message == nil || updates[0].Message.Text != "https://example.com" {
		t.Fatalf("unexpected update payload: %+v", updates[0])
	}
}

func TestSendMessageReturnsTelegramAPIDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":false,"error_code":400,"description":"chat not found"}`))
	}))
	t.Cleanup(server.Close)

	client := New(server.URL, "test-token")

	err := client.SendMessage(context.Background(), 77, "hello")
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "chat not found") {
		t.Fatalf("error = %q, want description", err)
	}
}
