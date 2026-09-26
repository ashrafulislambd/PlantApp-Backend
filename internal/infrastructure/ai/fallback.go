package ai

import (
	"context"
	"log"
	"time"

	"myplantpal-backend/internal/domain/chat"
	"myplantpal-backend/internal/domain/diagnosis"
)

const providerTimeout = 20 * time.Second

// ChatProviderEntry pairs a chat.ReplyProvider with a name used only for
// fallback logging (the provider's own result already carries its
// aiprovider.Name on success).
type ChatProviderEntry struct {
	Name     string
	Provider chat.ReplyProvider
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
	var lastErr error
	for _, e := range f.entries {
		attemptCtx, cancel := context.WithTimeout(ctx, providerTimeout)
		result, err := e.Provider.Reply(attemptCtx, history, userMessage, lang)
		cancel()
		if err == nil {
			return result, nil
		}
		log.Printf("ai: %s provider failed, falling back: %v", e.Name, err)
		lastErr = err
	}
	return chat.ReplyResult{}, lastErr
}

// DiagnosisProviderEntry is the diagnosis.Provider equivalent of
// ChatProviderEntry.
type DiagnosisProviderEntry struct {
	Name     string
	Provider diagnosis.Provider
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
	var lastErr error
	for _, e := range f.entries {
		attemptCtx, cancel := context.WithTimeout(ctx, providerTimeout)
		result, err := e.Provider.Analyze(attemptCtx, imageData)
		cancel()
		if err == nil {
			return result, nil
		}
		log.Printf("ai: %s provider failed, falling back: %v", e.Name, err)
		lastErr = err
	}
	return diagnosis.AnalysisResult{}, lastErr
}
