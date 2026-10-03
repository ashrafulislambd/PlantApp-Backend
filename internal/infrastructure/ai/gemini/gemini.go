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

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/aiprovider"
	"plantpal-backend/internal/domain/chat"
	"plantpal-backend/internal/domain/diagnosis"
	"plantpal-backend/internal/domain/plant"
	"plantpal-backend/internal/infrastructure/ai/sse"
)

const (
	defaultModel = "gemini-3.6-flash"
	apiBase      = "https://generativelanguage.googleapis.com/v1beta/models"
)

// Client calls the Gemini API. apiKey is sent via the x-goog-api-key
// header (not a query param).
type Client struct {
	apiKey  string
	model   string
	http    *http.Client
	baseURL string
	// streamHTTP has no overall Timeout (a streamed reply can outlast the
	// 30s a normal call gets); the caller's context bounds the stream.
	streamHTTP *http.Client
}

// New creates a Gemini client. An empty model falls back to
// "gemini-3.6-flash", a currently available multimodal Flash model.
func New(apiKey, model string) *Client {
	if model == "" {
		model = defaultModel
	}
	return &Client{
		apiKey:     apiKey,
		model:      model,
		http:       &http.Client{Timeout: 30 * time.Second},
		baseURL:    apiBase,
		streamHTTP: &http.Client{},
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
{
  "issue": "<short description of the problem in English>",
  "cure": "<short, actionable treatment in English>",
  "confidence": "High|Moderate|Low",
  "severity": "Mild|Moderate|Severe",
  "fertilizer": "<a short fertilizer or nutrient suggestion, or empty string if not applicable>",
  "issueBn": "<short description of the problem translated into natural Bengali/à¦¬à¦¾à¦‚à¦²à¦¾>",
  "cureBn": "<short, actionable treatment translated into natural Bengali/à¦¬à¦¾à¦‚à¦²à¦¾>"
}
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
	return fmt.Sprintf("%s/%s:generateContent", c.baseURL, c.model)
}

// streamEndpoint is the SSE variant: alt=sse makes Gemini send each partial
// response as a "data: {json}" event instead of one JSON array.
func (c *Client) streamEndpoint() string {
	return fmt.Sprintf("%s/%s:streamGenerateContent?alt=sse", c.baseURL, c.model)
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

	if resp.StatusCode == http.StatusTooManyRequests {
		return "", fmt.Errorf("gemini: status 429: %w", apperr.ErrRateLimited)
	}
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

// chatRequest assembles the conversation, prior turns and system prompt.
func chatRequest(history []*chat.Message, userMessage, lang string) geminiRequest {
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

	systemPrompt := chatSystemPrompt
	if lang == "bn" {
		systemPrompt += " You MUST answer the user in Bengali (à¦¬à¦¾à¦‚à¦²à¦¾). All advice, plant care tips, and explanations must be written in natural, fluent Bengali."
	}
	return geminiRequest{
		Contents:          contents,
		SystemInstruction: &geminiSystemInstruction{Parts: []geminiPart{{Text: systemPrompt}}},
	}
}

// Reply implements chat.ReplyProvider.
func (c *Client) Reply(ctx context.Context, history []*chat.Message, userMessage string, lang string) (chat.ReplyResult, error) {
	text, err := c.do(ctx, chatRequest(history, userMessage, lang))
	if err != nil {
		return chat.ReplyResult{}, err
	}
	return chat.ReplyResult{Text: text, Provider: aiprovider.Gemini}, nil
}

// ReplyStream implements chat.StreamProvider via streamGenerateContent with
// alt=sse. Each partial candidate's text is passed to onDelta immediately.
func (c *Client) ReplyStream(ctx context.Context, history []*chat.Message, userMessage string, lang string, onDelta func(string) error) (chat.ReplyResult, error) {
	body, err := json.Marshal(chatRequest(history, userMessage, lang))
	if err != nil {
		return chat.ReplyResult{}, fmt.Errorf("gemini: marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.streamEndpoint(), bytes.NewReader(body))
	if err != nil {
		return chat.ReplyResult{}, fmt.Errorf("gemini: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.streamHTTP.Do(httpReq)
	if err != nil {
		return chat.ReplyResult{}, fmt.Errorf("gemini: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return chat.ReplyResult{}, fmt.Errorf("gemini: status 429: %w", apperr.ErrRateLimited)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return chat.ReplyResult{}, fmt.Errorf("gemini: status %d", resp.StatusCode)
	}

	var sb strings.Builder
	result := func() chat.ReplyResult {
		return chat.ReplyResult{Text: sb.String(), Provider: aiprovider.Gemini}
	}
	err = sse.Read(resp.Body, func(data string) error {
		var gr geminiResponse
		if err := json.Unmarshal([]byte(data), &gr); err != nil {
			return fmt.Errorf("gemini: decode stream chunk: %w", err)
		}
		if len(gr.Candidates) == 0 {
			return nil
		}
		var piece strings.Builder
		for _, p := range gr.Candidates[0].Content.Parts {
			piece.WriteString(p.Text)
		}
		if piece.Len() == 0 {
			return nil
		}
		sb.WriteString(piece.String())
		return onDelta(piece.String())
	})
	if err != nil {
		return result(), fmt.Errorf("gemini: stream: %w", err)
	}
	if strings.TrimSpace(sb.String()) == "" {
		return chat.ReplyResult{}, fmt.Errorf("gemini: empty response text")
	}
	return result(), nil
}

type diagnosisJSON struct {
	Issue      string `json:"issue"`
	Cure       string `json:"cure"`
	Confidence string `json:"confidence,omitempty"`
	Severity   string `json:"severity,omitempty"`
	Fertilizer string `json:"fertilizer,omitempty"`
	IssueBn    string `json:"issueBn,omitempty"`
	CureBn     string `json:"cureBn,omitempty"`
}

// Analyze implements diagnosis.Provider.
func (c *Client) Analyze(ctx context.Context, imageData []byte, note string) (diagnosis.AnalysisResult, error) {
	mimeType := http.DetectContentType(imageData)
	encoded := base64.StdEncoding.EncodeToString(imageData)

	raw, err := c.do(ctx, geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: diagnosisPrompt + diagnosis.NoteHint(note)},
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

	return diagnosis.AnalysisResult{
		Issue:      dj.Issue,
		Cure:       dj.Cure,
		Confidence: dj.Confidence,
		Severity:   dj.Severity,
		Fertilizer: dj.Fertilizer,
		IssueBn:    dj.IssueBn,
		CureBn:     dj.CureBn,
		Provider:   aiprovider.Gemini,
	}, nil
}

type identifyJSON struct {
	Species               string `json:"species"`
	SuggestedNickname     string `json:"suggestedNickname"`
	Location              string `json:"location"`
	Sunlight              string `json:"sunlight"`
	WateringFrequencyDays int    `json:"wateringFrequencyDays"`
	WaterAmountMl         int    `json:"waterAmountMl"`
	Health                int    `json:"health"`
	CareTips              string `json:"careTips"`
}

const identifyPrompt = `You are an expert botanist and horticulturist. Look at this photo of a plant and identify the plant species, give a cute/friendly suggested nickname (such as Monty, Leafy, Spiky, Sunny, etc.), recommend suitable indoor/outdoor location, sunlight requirements (e.g., Bright indirect light, Low light, Direct sunlight), watering interval in days (an integer), watering amount in ml (an integer), detected health score (integer from 0 to 100), and concise care tips.
Reply ONLY as valid JSON with exactly these keys:
{
  "species": "<plant common name and scientific name in English>",
  "suggestedNickname": "<a cute, charming nickname for this plant>",
  "location": "<recommended placement, e.g. Indoor - Living Room, Balcony, etc.>",
  "sunlight": "<sunlight requirement, e.g. Bright indirect sunlight>",
  "wateringFrequencyDays": <integer, e.g. 7>,
  "waterAmountMl": <integer, e.g. 250>,
  "health": <integer from 0 to 100>,
  "careTips": "<short practical care tips>"
}
Do not add any text outside the JSON object.`

// Identify implements plant.Identifier.
func (c *Client) Identify(ctx context.Context, imageData []byte) (plant.IdentificationResult, error) {
	mimeType := http.DetectContentType(imageData)
	encoded := base64.StdEncoding.EncodeToString(imageData)

	raw, err := c.do(ctx, geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: identifyPrompt},
					{InlineData: &geminiInlineData{MimeType: mimeType, Data: encoded}},
				},
			},
		},
		GenerationConfig: &geminiGenerationConfig{ResponseMimeType: "application/json"},
	})
	if err != nil {
		return plant.IdentificationResult{}, err
	}

	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var ij identifyJSON
	if err := json.Unmarshal([]byte(raw), &ij); err != nil {
		return plant.IdentificationResult{}, fmt.Errorf("gemini: parse identify JSON %q: %w", raw, err)
	}

	return plant.IdentificationResult{
		Species:               ij.Species,
		SuggestedNickname:     ij.SuggestedNickname,
		Location:              ij.Location,
		Sunlight:              ij.Sunlight,
		WateringFrequencyDays: ij.WateringFrequencyDays,
		WaterAmountMl:         ij.WaterAmountMl,
		Health:                ij.Health,
		CareTips:              ij.CareTips,
	}, nil
}
