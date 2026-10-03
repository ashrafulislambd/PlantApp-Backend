package chat

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTitleFromContent(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"short", "Why are my leaves yellow?", "Why are my leaves yellow?"},
		{"collapses whitespace", "  hello \n\n  world\t!", "hello world !"},
		{"empty", "   ", ""},
	}
	for _, c := range cases {
		if got := TitleFromContent(c.in); got != c.want {
			t.Errorf("%s: TitleFromContent(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

func TestTitleFromContent_TruncatesOnRuneBoundary(t *testing.T) {
	long := strings.Repeat("পাতা ", 30) // Bengali: multi-byte runes
	got := TitleFromContent(long)
	if !utf8.ValidString(got) {
		t.Fatalf("title is not valid UTF-8: %q", got)
	}
	if n := utf8.RuneCountInString(got); n > MaxTitleRunes+1 {
		t.Errorf("title has %d runes, want <= %d", n, MaxTitleRunes+1)
	}
	if !strings.HasSuffix(got, "\u2026") {
		t.Errorf("title %q should end with an ellipsis", got)
	}
}
