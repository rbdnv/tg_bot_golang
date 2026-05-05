package telegram

import "testing"

func TestCommandName(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		want   string
		wantOK bool
	}{
		{name: "simple command", text: "/rnd", want: "/rnd", wantOK: true},
		{name: "command with mention", text: "/help@TestBot", want: "/help", wantOK: true},
		{name: "command with payload", text: "/start deep-link-payload", want: "/start", wantOK: true},
		{name: "command with extra spaces", text: "  /RND   ", want: "/rnd", wantOK: true},
		{name: "url is not command", text: "https://example.com", wantOK: false},
		{name: "empty text", text: "   ", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := commandName(tt.text)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}

			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
