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

const (
	// streamFirstTokenTimeout bounds how long a provider may stay silent
	// before its first piece of text; past it the next provider is tried.
	streamFirstTokenTimeout = 15 * time.Second
	// streamTotalTimeout bounds a whole streamed reply.
	streamTotalTimeout = 3 * time.Minute
)

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

// ReplyStream tries each entry in order like Reply, but streams. Falling
// back is only possible while nothing has been sent to the caller: once a
// provider has delivered text, a failure is returned as-is (with the partial
// result) instead of restarting the answer with a different provider.
//
// Entries that do not implement chat.StreamProvider are called through
// Reply and delivered as a single piece.
func (f *FallbackChatProvider) ReplyStream(ctx context.Context, history []*chat.Message, userMessage string, lang string, onDelta func(string) error) (chat.ReplyResult, error) {
	if len(f.entries) == 0 {
		return chat.ReplyResult{}, fmt.Errorf("ai: no providers configured")
	}

	delivered := false
	emit := func(piece string) error {
		delivered = true
		return onDelta(piece)
	}

	var lastErr error
	rateLimited := false
	for _, e := range f.entries {
		if e.Placeholder && rateLimited {
			return chat.ReplyResult{}, apperr.ErrRateLimited
		}

		result, err := f.streamOne(ctx, e, history, userMessage, lang, emit)
		if err == nil {
			return result, nil
		}
		if delivered || ctx.Err() != nil {
			// Too late to fall back (the user already saw text), or the
			// caller is gone.
			return result, err
		}
		log.Printf("ai: %s provider failed, falling back: %v", e.Name, err)
		if errors.Is(err, apperr.ErrRateLimited) {
			rateLimited = true
		}
		lastErr = err
	}
	return chat.ReplyResult{}, lastErr
}

// streamOne runs a single entry under the time-to-first-token and total
// deadlines.
func (f *FallbackChatProvider) streamOne(ctx context.Context, e ChatProviderEntry, history []*chat.Message, userMessage, lang string, emit func(string) error) (chat.ReplyResult, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, streamTotalTimeout)
	defer cancel()

	sp, ok := e.Provider.(chat.StreamProvider)
	if !ok {
		singleCtx, singleCancel := context.WithTimeout(ctx, providerTimeout)
		defer singleCancel()
		res, err := e.Provider.Reply(singleCtx, history, userMessage, lang)
		if err != nil {
			return chat.ReplyResult{}, err
		}
		if err := emit(res.Text); err != nil {
			return res, err
		}
		return res, nil
	}

	firstToken := time.AfterFunc(streamFirstTokenTimeout, cancel)
	defer firstToken.Stop()
	return sp.ReplyStream(attemptCtx, history, userMessage, lang, func(piece string) error {
		firstToken.Stop()
		return emit(piece)
	})
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

func (f *FallbackDiagnosisProvider) Analyze(ctx context.Context, imageData []byte, note string) (diagnosis.AnalysisResult, error) {
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
		result, err := e.Provider.Analyze(attemptCtx, imageData, note)
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
