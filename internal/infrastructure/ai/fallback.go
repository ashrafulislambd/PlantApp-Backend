package ai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/chat"
	"plantpal-backend/internal/domain/diagnosis"
)

const providerTimeout = 20 * time.Second

// ChatProviderEntry pairs a chat.ReplyProvider with a name used only for
// fallback logging (the provider's own result already carries its
// aiprovider.Name on success).
type ChatProviderEntry struct {
	Name     string
	Provider chat.ReplyProvider
	// Placeholder marks a canned last-resort provider (the mock). If a real
	// provider before it was rate limited, the chain returns
	// apperr.ErrRateLimited instead of masking it with placeholder output.
	Placeholder bool
}

// FallbackChatProvider tries each entry in order, logging a warning and
// moving on when one fails, returning the first success. The last entry is
// expected to be a provider that cannot fail (the mock), so the chain
// always resolves.
type FallbackChatProvider struct {
	entries []ChatProviderEntry
}

// NewFallbackChatProvider builds a fallback chain tried in the given order.
func NewFallbackChatProvider(entries ...ChatProviderEntry) *FallbackChatProvider {
	return &FallbackChatProvider{entries: entries}
}

func (f *FallbackChatProvider) Reply(ctx context.Context, history []*chat.Message, userMessage string, lang string) (chat.ReplyResult, error) {
	if len(f.entries) == 0 {
		return chat.ReplyResult{}, fmt.Errorf("ai: no providers configured")
	}

	var lastErr error
	rateLimited := false
	for _, e := range f.entries {
		if e.Placeholder && rateLimited {
			return chat.ReplyResult{}, apperr.ErrRateLimited
		}
		attemptCtx, cancel := context.WithTimeout(ctx, providerTimeout)
		result, err := e.Provider.Reply(attemptCtx, history, userMessage, lang)
		cancel()
		if err == nil {
			return result, nil
		}
		log.Printf("ai: %s provider failed, falling back: %v", e.Name, err)
		if errors.Is(err, apperr.ErrRateLimited) {
			rateLimited = true
		}
		lastErr = err
	}
	return chat.ReplyResult{}, lastErr
}

// DiagnosisProviderEntry is the diagnosis.Provider equivalent of
// ChatProviderEntry.
type DiagnosisProviderEntry struct {
	Name        string
	Provider    diagnosis.Provider
	Placeholder bool // see ChatProviderEntry.Placeholder
}

// FallbackDiagnosisProvider is the diagnosis.Provider equivalent of
// FallbackChatProvider.
type FallbackDiagnosisProvider struct {
	entries []DiagnosisProviderEntry
}

// NewFallbackDiagnosisProvider builds a fallback chain tried in the given
// order.
func NewFallbackDiagnosisProvider(entries ...DiagnosisProviderEntry) *FallbackDiagnosisProvider {
	return &FallbackDiagnosisProvider{entries: entries}
}

func (f *FallbackDiagnosisProvider) Analyze(ctx context.Context, imageData []byte) (diagnosis.AnalysisResult, error) {
	if len(f.entries) == 0 {
		return diagnosis.AnalysisResult{}, fmt.Errorf("ai: no providers configured")
	}

	var lastErr error
	rateLimited := false
	for _, e := range f.entries {
		if e.Placeholder && rateLimited {
			return diagnosis.AnalysisResult{}, apperr.ErrRateLimited
		}
		attemptCtx, cancel := context.WithTimeout(ctx, providerTimeout)
		result, err := e.Provider.Analyze(attemptCtx, imageData)
		cancel()
		if err == nil {
			return result, nil
		}
		log.Printf("ai: %s provider failed, falling back: %v", e.Name, err)
		if errors.Is(err, apperr.ErrRateLimited) {
			rateLimited = true
		}
		lastErr = err
	}
	return diagnosis.AnalysisResult{}, lastErr
}
