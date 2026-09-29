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

	_, err := provider.Analyze(context.Background(), nil)
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
