// Package gemini implements chat.ReplyProvider and diagnosis.Provider using
// Google's Gemini generateContent API, called directly over HTTP to keep
// the backend free of external dependencies.
package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"myplantpal-backend/internal/domain/aiprovider"
	"myplantpal-backend/internal/domain/chat"
	"myplantpal-backend/internal/domain/diagnosis"
)

const (
	defaultModel = "gemini-3.6-flash"
	apiBase      = "https://generativelanguage.googleapis.com/v1beta/models"
)

// Client calls the Gemini API. apiKey is sent via the x-goog-api-key
// header (not a query param).
type Client struct {
	apiKey string
	model  string
	http   *http.Client
}

// New creates a Gemini client. An empty model falls back to
// "gemini-3.6-flash", a currently available multimodal Flash model.
func New(apiKey, model string) *Client {
	if model == "" {
		model = defaultModel
	}
	return &Client{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inline_data,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiSystemInstruction struct {
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string `json:"responseMimeType,omitempty"`
}

type geminiRequest struct {
	Contents          []geminiContent          `json:"contents"`
	SystemInstruction *geminiSystemInstruction `json:"system_instruction,omitempty"`
	GenerationConfig  *geminiGenerationConfig  `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

const chatSystemPrompt = "You are the AI Doctor's chat assistant inside the PlantPal app. " +
	"Answer the user's plant-care question concisely, warmly, and practically."

const diagnosisPrompt = `You are a plant pathologist. Look at this photo of a plant and identify ` +
	`the single most likely issue (disease, pest, or nutrient deficiency) and a practical cure.
Reply ONLY as valid JSON with exactly these keys:
{"issue": "<short description of the problem>", "cure": "<short, actionable treatment>"}
Do not add any text outside the JSON object.`

// roleToGemini remaps the domain's chat roles onto Gemini's, which uses
// "model" (not "assistant") for prior assistant turns.
func roleToGemini(r chat.Role) string {
	if r == chat.RoleAssistant {
		return "model"
	}
	return "user"
}

func (c *Client) endpoint() string {
	return fmt.Sprintf("%s/%s:generateContent", apiBase, c.model)
}

// do sends a generateContent request and returns the concatenated text of
// the first candidate. Any non-2xx status or decode failure returns an
// error so the fallback chain moves on to the next provider.
func (c *Client) do(ctx context.Context, req geminiRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("gemini: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("gemini: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("gemini: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini: status %d", resp.StatusCode)
	}

	var gr geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return "", fmt.Errorf("gemini: decode response: %w", err)
	}
	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: no candidates returned")
	}

	var sb strings.Builder
	for _, p := range gr.Candidates[0].Content.Parts {
		sb.WriteString(p.Text)
	}
	text := strings.TrimSpace(sb.String())
	if text == "" {
		return "", fmt.Errorf("gemini: empty response text")
	}
	return text, nil
}

// Reply implements chat.ReplyProvider.
func (c *Client) Reply(ctx context.Context, history []*chat.Message, userMessage string) (chat.ReplyResult, error) {
	contents := make([]geminiContent, 0, len(history)+1)
	for _, m := range history {
		contents = append(contents, geminiContent{
			Role:  roleToGemini(m.Role),
			Parts: []geminiPart{{Text: m.Content}},
		})
	}
	contents = append(contents, geminiContent{
		Role:  "user",
		Parts: []geminiPart{{Text: userMessage}},
	})

	text, err := c.do(ctx, geminiRequest{
		Contents:          contents,
		SystemInstruction: &geminiSystemInstruction{Parts: []geminiPart{{Text: chatSystemPrompt}}},
	})
	if err != nil {
		return chat.ReplyResult{}, err
	}
	return chat.ReplyResult{Text: text, Provider: aiprovider.Gemini}, nil
}

type diagnosisJSON struct {
	Issue string `json:"issue"`
	Cure  string `json:"cure"`
}

// Analyze implements diagnosis.Provider.
func (c *Client) Analyze(ctx context.Context, imageData []byte) (diagnosis.AnalysisResult, error) {
	mimeType := http.DetectContentType(imageData)
	encoded := base64.StdEncoding.EncodeToString(imageData)

	raw, err := c.do(ctx, geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: diagnosisPrompt},
					{InlineData: &geminiInlineData{MimeType: mimeType, Data: encoded}},
				},
			},
		},
		GenerationConfig: &geminiGenerationConfig{ResponseMimeType: "application/json"},
	})
	if err != nil {
		return diagnosis.AnalysisResult{}, err
	}

	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var dj diagnosisJSON
	if err := json.Unmarshal([]byte(raw), &dj); err != nil {
		return diagnosis.AnalysisResult{}, fmt.Errorf("gemini: parse diagnosis JSON %q: %w", raw, err)
	}

	return diagnosis.AnalysisResult{Issue: dj.Issue, Cure: dj.Cure, Provider: aiprovider.Gemini}, nil
}
