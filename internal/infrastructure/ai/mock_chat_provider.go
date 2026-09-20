package ai

import (
	"context"
	"fmt"

	"myplantpal-backend/internal/domain/chat"
)

// MockChatReplyProvider returns a canned acknowledgement referencing the
// user's message, so the AI Chat Box has something to display end to end.
type MockChatReplyProvider struct{}

func NewMockChatReplyProvider() *MockChatReplyProvider {
	return &MockChatReplyProvider{}
}

func (p *MockChatReplyProvider) Reply(_ context.Context, _ []*chat.Message, userMessage string) (string, error) {
	return fmt.Sprintf(
		"Thanks for asking about %q. An expert reply isn't wired up yet — check the Fertilizer or AI Doctor screens for guidance in the meantime.",
		userMessage,
	), nil
}
