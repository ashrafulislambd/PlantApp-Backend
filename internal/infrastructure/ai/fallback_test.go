package ai

import (
	"context"
	"errors"
	"strings"
	"testing"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/chat"
)

func TestFallbackChatProvider_NoEntriesReturnsError(t *testing.T) {
	provider := NewFallbackChatProvider()

	_, err := provider.Reply(context.Background(), nil, "hello", "en")
	if err == nil {
		t.Fatal("Reply() error = nil, want error when no providers are configured")
	}
	if !strings.Contains(err.Error(), "no providers configured") {
		t.Fatalf("Reply() error = %v, want message containing %q", err, "no providers configured")
	}
}

func TestFallbackDiagnosisProvider_NoEntriesReturnsError(t *testing.T) {
	provider := NewFallbackDiagnosisProvider()

	_, err := provider.Analyze(context.Background(), nil, "")
	if err == nil {
		t.Fatal("Analyze() error = nil, want error when no providers are configured")
	}
	if !strings.Contains(err.Error(), "no providers configured") {
		t.Fatalf("Analyze() error = %v, want message containing %q", err, "no providers configured")
	}
}

type rateLimitedChat struct{}

func (rateLimitedChat) Reply(context.Context, []*chat.Message, string, string) (chat.ReplyResult, error) {
	return chat.ReplyResult{}, apperr.ErrRateLimited
}

type okChat struct{}

func (okChat) Reply(context.Context, []*chat.Message, string, string) (chat.ReplyResult, error) {
	return chat.ReplyResult{Text: "mock"}, nil
}

func TestFallbackChatProvider_RateLimitNotMaskedByPlaceholder(t *testing.T) {
	provider := NewFallbackChatProvider(
		ChatProviderEntry{Name: "real", Provider: rateLimitedChat{}},
		ChatProviderEntry{Name: "mock", Provider: okChat{}, Placeholder: true},
	)
	_, err := provider.Reply(context.Background(), nil, "hi", "en")
	if !errors.Is(err, apperr.ErrRateLimited) {
		t.Fatalf("Reply() error = %v, want ErrRateLimited", err)
	}
}

func TestFallbackChatProvider_PlaceholderStillUsedForOtherFailures(t *testing.T) {
	failing := chatFunc(func() (chat.ReplyResult, error) { return chat.ReplyResult{}, errors.New("boom") })
	provider := NewFallbackChatProvider(
		ChatProviderEntry{Name: "real", Provider: failing},
		ChatProviderEntry{Name: "mock", Provider: okChat{}, Placeholder: true},
	)
	res, err := provider.Reply(context.Background(), nil, "hi", "en")
	if err != nil || res.Text != "mock" {
		t.Fatalf("Reply() = %v, %v; want mock reply", res, err)
	}
}

type chatFunc func() (chat.ReplyResult, error)

func (f chatFunc) Reply(context.Context, []*chat.Message, string, string) (chat.ReplyResult, error) {
	return f()
}

// streamEntry is a chat provider that streams scripted pieces then fails
// (or succeeds when err is nil).
type streamEntry struct {
	pieces []string
	err    error
	calls  *int
}

func (s streamEntry) Reply(context.Context, []*chat.Message, string, string) (chat.ReplyResult, error) {
	return chat.ReplyResult{}, errors.New("unused")
}

func (s streamEntry) ReplyStream(_ context.Context, _ []*chat.Message, _ string, _ string, onDelta func(string) error) (chat.ReplyResult, error) {
	if s.calls != nil {
		*s.calls++
	}
	text := ""
	for _, p := range s.pieces {
		text += p
		if err := onDelta(p); err != nil {
			return chat.ReplyResult{Text: text}, err
		}
	}
	return chat.ReplyResult{Text: text}, s.err
}

func collectStream(p *FallbackChatProvider) ([]string, chat.ReplyResult, error) {
	var got []string
	res, err := p.ReplyStream(context.Background(), nil, "hi", "en", func(s string) error {
		got = append(got, s)
		return nil
	})
	return got, res, err
}

func TestFallbackReplyStream_FallsBackWhenFailureIsBeforeFirstPiece(t *testing.T) {
	provider := NewFallbackChatProvider(
		ChatProviderEntry{Name: "bad", Provider: streamEntry{err: errors.New("boom")}},
		ChatProviderEntry{Name: "good", Provider: streamEntry{pieces: []string{"he", "llo"}}},
	)
	got, res, err := collectStream(provider)
	if err != nil || res.Text != "hello" || len(got) != 2 {
		t.Fatalf("got = %q, res = %+v, err = %v", got, res, err)
	}
}

func TestFallbackReplyStream_DoesNotRestartAfterTextWasDelivered(t *testing.T) {
	boom := errors.New("mid-stream failure")
	secondCalls := 0
	provider := NewFallbackChatProvider(
		ChatProviderEntry{Name: "flaky", Provider: streamEntry{pieces: []string{"par", "tial"}, err: boom}},
		ChatProviderEntry{Name: "other", Provider: streamEntry{pieces: []string{"never"}, calls: &secondCalls}},
	)
	got, res, err := collectStream(provider)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if res.Text != "partial" || len(got) != 2 {
		t.Errorf("got = %q, res = %+v; want the partial text kept", got, res)
	}
	if secondCalls != 0 {
		t.Errorf("second provider was called %d times; must not restart after output", secondCalls)
	}
}

func TestFallbackReplyStream_RateLimitNotMaskedByPlaceholder(t *testing.T) {
	provider := NewFallbackChatProvider(
		ChatProviderEntry{Name: "real", Provider: streamEntry{err: apperr.ErrRateLimited}},
		ChatProviderEntry{Name: "mock", Provider: streamEntry{pieces: []string{"canned"}}, Placeholder: true},
	)
	_, _, err := collectStream(provider)
	if !errors.Is(err, apperr.ErrRateLimited) {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
}

func TestFallbackReplyStream_NonStreamingEntryIsDeliveredAsOnePiece(t *testing.T) {
	provider := NewFallbackChatProvider(ChatProviderEntry{Name: "plain", Provider: okChat{}})
	got, res, err := collectStream(provider)
	if err != nil || res.Text != "mock" || len(got) != 1 || got[0] != "mock" {
		t.Fatalf("got = %q, res = %+v, err = %v", got, res, err)
	}
}

func TestFallbackReplyStream_NoEntriesReturnsError(t *testing.T) {
	_, _, err := collectStream(NewFallbackChatProvider())
	if err == nil || !strings.Contains(err.Error(), "no providers configured") {
		t.Fatalf("err = %v", err)
	}
}

func TestMockChatProvider_StreamsWholeReply(t *testing.T) {
	m := NewMockChatReplyProvider()
	var got []string
	res, err := m.ReplyStream(context.Background(), nil, "roses", "en", func(s string) error {
		got = append(got, s)
		return nil
	})
	if err != nil {
		t.Fatalf("ReplyStream() error = %v", err)
	}
	if strings.Join(got, "") != res.Text || len(got) < 2 {
		t.Errorf("pieces %q do not add up to %q", got, res.Text)
	}
}

func TestMockChatProvider_StopsWhenContextCancelled(t *testing.T) {
	m := NewMockChatReplyProvider()
	ctx, cancel := context.WithCancel(context.Background())
	n := 0
	_, err := m.ReplyStream(ctx, nil, "roses", "en", func(string) error {
		n++
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) || n != 1 {
		t.Errorf("err = %v, pieces = %d; want context.Canceled after 1", err, n)
	}
}
