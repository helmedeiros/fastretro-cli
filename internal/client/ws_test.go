package client

import "testing"

func TestExtractRoomCode_DirectCode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ABC-123-DEF", "ABC-123-DEF"},
		{"abc-123-def", "abc-123-def"},
		{"ROOM1", "ROOM1"},
		{"A1B2C", "A1B2C"},
	}

	for _, tt := range tests {
		got := extractRoomCode(tt.input)
		if got != tt.want {
			t.Errorf("extractRoomCode(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractRoomCode_FromURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"http://localhost:5173/#room=ABC-123-DEF", "ABC-123-DEF"},
		{"https://retro.example.com/#room=XYZ-789", "XYZ-789"},
		{"http://localhost:5173/#room=ROOMCODE", "ROOMCODE"},
	}

	for _, tt := range tests {
		got := extractRoomCode(tt.input)
		if got != tt.want {
			t.Errorf("extractRoomCode(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractRoomCode_Invalid(t *testing.T) {
	tests := []string{
		"",
		"AB",
		"!!!",
		"http://localhost:5173/",
		"http://localhost:5173/#other=ABC",
	}

	for _, input := range tests {
		got := extractRoomCode(input)
		if got != "" {
			t.Errorf("extractRoomCode(%q) = %q, want empty", input, got)
		}
	}
}

func TestExtractRoomCode_WithWhitespace(t *testing.T) {
	got := extractRoomCode("  ABC-123-DEF  ")
	if got != "ABC-123-DEF" {
		t.Errorf("got %q, want 'ABC-123-DEF'", got)
	}
}

func TestIsRoomCode(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"ABC-123-DEF", true},
		{"abcde", true},
		{"12345", true},
		{"A-B-C-D-E", true},
		{"AB", false},
		{"", false},
		{"ABC DEF", false},
		{"ABC@DEF", false},
	}

	for _, tt := range tests {
		got := isRoomCode(tt.input)
		if got != tt.want {
			t.Errorf("isRoomCode(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestToWSURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"http://localhost:5173", "ws://localhost:5173"},
		{"https://retro.example.com", "wss://retro.example.com"},
		{"http://localhost:5173/", "ws://localhost:5173"},
		{"localhost:5173", "ws://localhost:5173"},
		{"ws://already-ws:5173", "ws://already-ws:5173"},
		{"wss://already-wss:5173", "wss://already-wss:5173"},
	}

	for _, tt := range tests {
		got := toWSURL(tt.input)
		if got != tt.want {
			t.Errorf("toWSURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestShareURL(t *testing.T) {
	c := &Client{RoomCode: "ABC-123-DEF"}
	got := c.ShareURL("http://localhost:5173")
	want := "http://localhost:5173/#room=ABC-123-DEF"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestShareURL_TrailingSlash(t *testing.T) {
	c := &Client{RoomCode: "ABC-123"}
	got := c.ShareURL("http://localhost:5173/")
	want := "http://localhost:5173/#room=ABC-123"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
