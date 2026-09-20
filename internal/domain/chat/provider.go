package chat

import "context"

// ReplyProvider abstracts the AI backend that generates assistant replies.
// The in-memory mock in internal/infrastructure/ai satisfies it for now;
// swap it for a real Gemini/Groq-backed implementation later.
type ReplyProvider interface {
	Reply(ctx context.Context, history []*Message, userMessage string) (string, error)
}
