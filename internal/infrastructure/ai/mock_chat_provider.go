package ai

import (
	"context"
	"fmt"

	"myplantpal-backend/internal/domain/aiprovider"
	"myplantpal-backend/internal/domain/chat"
)

// MockChatReplyProvider returns a canned acknowledgement referencing the
// user's message. It is the last resort in the fallback chain, used when
// no real AI provider is configured or all of them failed.
type MockChatReplyProvider struct{}

func NewMockChatReplyProvider() *MockChatReplyProvider {
	return &MockChatReplyProvider{}
}

func (p *MockChatReplyProvider) Reply(_ context.Context, _ []*chat.Message, userMessage string) (chat.ReplyResult, error) {
	return chat.ReplyResult{
		Text: fmt.Sprintf(
			"Thanks for asking about %q. An expert reply isn't wired up yet — check the Fertilizer or AI Doctor screens for guidance in the meantime.",
			userMessage,
		),
		Provider: aiprovider.Mock,
	}, nil
}
