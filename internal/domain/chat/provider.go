package chat

import (
	"context"

	"plantpal-backend/internal/domain/aiprovider"
)

// ReplyResult is what an AI provider returns for a chat turn: the reply
// text plus which provider actually produced it.
type ReplyResult struct {
	Text     string
	Provider aiprovider.Name
}

// ReplyProvider abstracts the AI backend that generates assistant replies.
// The in-memory mock in internal/infrastructure/ai satisfies it for now;
// swap it for a real Gemini/Groq-backed implementation later.
type ReplyProvider interface {
	Reply(ctx context.Context, history []*Message, userMessage string, lang string) (ReplyResult, error)
}

// StreamProvider is implemented by providers that can deliver a reply
// incrementally. onDelta is called with each new piece of text as soon as
// it arrives; if it returns an error (for example the client went away)
// the provider must stop and return that error. The returned ReplyResult
// carries the full text.
//
// If the stream fails after some text was delivered, the error is returned
// together with a ReplyResult holding the partial text, so callers can keep
// what the user already saw.
type StreamProvider interface {
	ReplyStream(ctx context.Context, history []*Message, userMessage string, lang string, onDelta func(string) error) (ReplyResult, error)
}
