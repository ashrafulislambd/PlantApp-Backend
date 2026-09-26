package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"myplantpal-backend/internal/domain/aiprovider"
	"myplantpal-backend/internal/domain/chat"
)

const (
	chatCompletionsEndpoint = "https://api.groq.com/openai/v1/chat/completions"
	defaultChatModel        = "llama-3.3-70b-versatile"
)

const chatSystemPrompt = "You are the AI Doctor's chat assistant inside the PlantPal app. " +
	"Answer the user's plant-care question concisely, warmly, and practically."

// ChatProvider implements chat.ReplyProvider using Groq's
// OpenAI-compatible chat completions API. Same call style as the existing
// price-refresher in this package.
type ChatProvider struct {
	apiKey string
	model  string
	http   *http.Client
}

// NewChatProvider creates a Groq chat provider. An empty model falls back
// to "llama-3.3-70b-versatile".
func NewChatProvider(apiKey, model string) *ChatProvider {
	if model == "" {
		model = defaultChatModel
	}
	return &ChatProvider{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model    string                  `json:"model"`
	Messages []chatCompletionMessage `json:"messages"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatCompletionMessage `json:"message"`
	} `json:"choices"`
}

// Reply implements chat.ReplyProvider. Domain roles ("user"/"assistant")
// match Groq's OpenAI-compatible roles directly — no remapping needed.
func (p *ChatProvider) Reply(ctx context.Context, history []*chat.Message, userMessage string, lang string) (chat.ReplyResult, error) {
	systemPrompt := chatSystemPrompt
	if lang == "bn" {
		systemPrompt += " You MUST answer the user in Bengali (বাংলা). All advice, plant care tips, and explanations must be written in natural, fluent Bengali."
	}
	messages := make([]chatCompletionMessage, 0, len(history)+2)
	messages = append(messages, chatCompletionMessage{Role: "system", Content: systemPrompt})
	for _, m := range history {
		messages = append(messages, chatCompletionMessage{Role: string(m.Role), Content: m.Content})
	}
	messages = append(messages, chatCompletionMessage{Role: "user", Content: userMessage})

	body, err := json.Marshal(chatCompletionRequest{Model: p.model, Messages: messages})
	if err != nil {
		return chat.ReplyResult{}, fmt.Errorf("groq: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatCompletionsEndpoint, bytes.NewReader(body))
	if err != nil {
		return chat.ReplyResult{}, fmt.Errorf("groq: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return chat.ReplyResult{}, fmt.Errorf("groq: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return chat.ReplyResult{}, fmt.Errorf("groq: status %d", resp.StatusCode)
	}

	var cr chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return chat.ReplyResult{}, fmt.Errorf("groq: decode response: %w", err)
	}
	if len(cr.Choices) == 0 {
		return chat.ReplyResult{}, fmt.Errorf("groq: no choices returned")
	}

	text := strings.TrimSpace(cr.Choices[0].Message.Content)
	if text == "" {
		return chat.ReplyResult{}, fmt.Errorf("groq: empty response text")
	}
	return chat.ReplyResult{Text: text, Provider: aiprovider.Groq}, nil
}
