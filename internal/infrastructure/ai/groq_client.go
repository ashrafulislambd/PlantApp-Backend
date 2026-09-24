package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const groqChatCompletionsURL = "https://api.groq.com/openai/v1/chat/completions"

// groqClient is a minimal client for Groq's OpenAI-compatible Chat
// Completions API — no SDK, just net/http, matching this backend's
// zero-dependency style (see go.mod).
type groqClient struct {
	apiKey string
	http   *http.Client
}

func newGroqClient(apiKey string) *groqClient {
	return &groqClient{apiKey: apiKey, http: &http.Client{Timeout: 30 * time.Second}}
}

type groqMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type groqTextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type groqImageContent struct {
	Type     string       `json:"type"`
	ImageURL groqImageURL `json:"image_url"`
}

type groqImageURL struct {
	URL string `json:"url"`
}

type groqResponseFormat struct {
	Type string `json:"type"`
}

type groqChatRequest struct {
	Model          string              `json:"model"`
	Messages       []groqMessage       `json:"messages"`
	Temperature    float64             `json:"temperature"`
	ResponseFormat *groqResponseFormat `json:"response_format,omitempty"`
}

type groqChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// complete sends a chat-completion request and returns the first choice's
// message content.
func (c *groqClient) complete(ctx context.Context, req groqChatRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("groq: encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, groqChatCompletionsURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("groq: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("groq: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("groq: read response: %w", err)
	}

	var parsed groqChatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("groq: decode response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return "", fmt.Errorf("groq: api error (%d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return "", fmt.Errorf("groq: api error (%d)", resp.StatusCode)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("groq: empty response")
	}
	return parsed.Choices[0].Message.Content, nil
}
