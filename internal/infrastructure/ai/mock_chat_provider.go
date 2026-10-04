package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"plantpal-backend/internal/domain/aiprovider"
	"plantpal-backend/internal/domain/chat"
)

// MockChatReplyProvider returns a canned acknowledgement referencing the
// user's message. It is the last resort in the fallback chain, used when
// no real AI provider is configured or all of them failed.
type MockChatReplyProvider struct{}

func NewMockChatReplyProvider() *MockChatReplyProvider {
	return &MockChatReplyProvider{}
}

func (p *MockChatReplyProvider) Reply(_ context.Context, _ []*chat.Message, userMessage string, lang string) (chat.ReplyResult, error) {
	return chat.ReplyResult{
		Text: fmt.Sprintf(
			"Thanks for asking about %q. An expert reply isn't wired up yet — check the Fertilizer or AI Doctor screens for guidance in the meantime.",
			userMessage,
		),
		Provider: aiprovider.Mock,
	}, nil
}

// mockStreamDelay paces the canned reply so the streaming path can be
// exercised end to end without a real AI key.
const mockStreamDelay = 20 * time.Millisecond

// ReplyStream implements chat.StreamProvider by emitting the canned reply
// word by word.
func (p *MockChatReplyProvider) ReplyStream(ctx context.Context, history []*chat.Message, userMessage string, lang string, onDelta func(string) error) (chat.ReplyResult, error) {
	full, _ := p.Reply(ctx, history, userMessage, lang)

	words := strings.SplitAfter(full.Text, " ")
	var sent strings.Builder
	for _, w := range words {
		if w == "" {
			continue
		}
		select {
		case <-ctx.Done():
			return chat.ReplyResult{Text: sent.String(), Provider: aiprovider.Mock}, ctx.Err()
		case <-time.After(mockStreamDelay):
		}
		sent.WriteString(w)
		if err := onDelta(w); err != nil {
			return chat.ReplyResult{Text: sent.String(), Provider: aiprovider.Mock}, err
		}
	}
	return full, nil
}
