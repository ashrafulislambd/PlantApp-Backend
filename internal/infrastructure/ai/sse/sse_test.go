package sse

import (
	"errors"
	"strings"
	"testing"
)

func collect(t *testing.T, in string) []string {
	t.Helper()
	var got []string
	if err := Read(strings.NewReader(in), func(d string) error {
		got = append(got, d)
		return nil
	}); err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	return got
}

func TestRead_SplitsEventsAndHandlesCRLF(t *testing.T) {
	got := collect(t, "data: one\r\n\r\ndata: two\n\n")
	if len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("got %q, want [one two]", got)
	}
}

func TestRead_JoinsMultilineData_IgnoresCommentsAndOtherFields(t *testing.T) {
	got := collect(t, ": keep-alive\nevent: x\nid: 7\ndata: a\ndata: b\n\n")
	if len(got) != 1 || got[0] != "a\nb" {
		t.Fatalf("got %q, want [a\\nb]", got)
	}
}

func TestRead_FlushesFinalEventWithoutTrailingBlankLine(t *testing.T) {
	got := collect(t, "data: last")
	if len(got) != 1 || got[0] != "last" {
		t.Fatalf("got %q, want [last]", got)
	}
}

func TestRead_ErrStopEndsQuietly(t *testing.T) {
	calls := 0
	err := Read(strings.NewReader("data: a\n\ndata: b\n\n"), func(string) error {
		calls++
		return ErrStop
	})
	if err != nil || calls != 1 {
		t.Fatalf("err = %v, calls = %d; want nil, 1", err, calls)
	}
}

func TestRead_PropagatesCallbackError(t *testing.T) {
	boom := errors.New("boom")
	err := Read(strings.NewReader("data: a\n\n"), func(string) error { return boom })
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}
