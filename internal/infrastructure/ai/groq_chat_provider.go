package ai

import (
	"context"
	"fmt"

	"myplantpal-backend/internal/domain/chat"
)

const chatSystemPrompt = `You are the MyPlantPal AI assistant: a friendly, concise plant-care helper.
Answer questions about watering, sunlight, fertilizing, pests, and general plant health.
Keep replies short (2-4 sentences) and practical. You are giving general guidance, not a
guaranteed diagnosis — if the user describes something serious (disease, pest infestation),
suggest they also try the AI Doctor photo scan for a closer look.`

// GroqChatReplyProvider generates AI Chat Box replies via Groq's chat
// completions API.
type GroqChatReplyProvider struct {
	client *groqClient
	model  string
}

func NewGroqChatReplyProvider(apiKey, model string) *GroqChatReplyProvider {
	return &GroqChatReplyProvider{client: newGroqClient(apiKey), model: model}
}

func (p *GroqChatReplyProvider) Reply(ctx context.Context, history []*chat.Message, userMessage string) (string, error) {
	messages := make([]groqMessage, 0, len(history)+2)
	messages = append(messages, groqMessage{Role: "system", Content: chatSystemPrompt})
	for _, m := range history {
		role := "user"
		if m.Role == chat.RoleAssistant {
			role = "assistant"
		}
		messages = append(messages, groqMessage{Role: role, Content: m.Content})
	}
	messages = append(messages, groqMessage{Role: "user", Content: userMessage})

	reply, err := p.client.complete(ctx, groqChatRequest{
		Model:       p.model,
		Messages:    messages,
		Temperature: 0.6,
	})
	if err != nil {
		return "", fmt.Errorf("groq chat: %w", err)
	}
	return reply, nil
}
