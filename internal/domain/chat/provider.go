package chat

import (
	"context"

	"myplantpal-backend/internal/domain/aiprovider"
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
	Reply(ctx context.Context, history []*Message, userMessage string) (ReplyResult, error)
}
