package groq

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
	"myplantpal-backend/internal/domain/diagnosis"
)

const defaultVisionModel = "qwen/qwen3.8-27b"

const diagnosisPrompt = `You are a plant pathologist. Look at this photo of a plant and identify ` +
	`the single most likely issue (disease, pest, or nutrient deficiency) and a practical cure.
Reply ONLY as valid JSON with exactly these keys:
{"issue": "<short description of the problem>", "cure": "<short, actionable treatment>"}
Do not add any text outside the JSON object.`

// DiagnosisProvider implements diagnosis.Provider using Groq's
// OpenAI-compatible, vision-capable chat completions API.
type DiagnosisProvider struct {
	apiKey string
	model  string
	http   *http.Client
}

// NewDiagnosisProvider creates a Groq vision provider. An empty model
// falls back to "qwen/qwen3.8-27b".
func NewDiagnosisProvider(apiKey, model string) *DiagnosisProvider {
	if model == "" {
		model = defaultVisionModel
	}
	return &DiagnosisProvider{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

type visionContentPart struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *visionImageURL `json:"image_url,omitempty"`
}

type visionImageURL struct {
	URL string `json:"url"`
}

type visionMessage struct {
	Role    string              `json:"role"`
	Content []visionContentPart `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type visionRequest struct {
	Model          string          `json:"model"`
	Messages       []visionMessage `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type visionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type diagnosisJSON struct {
	Issue string `json:"issue"`
	Cure  string `json:"cure"`
}

// Analyze implements diagnosis.Provider. Groq accepts base64 images inline
// (not just hosted URLs) via a data: URL. Note: the base64 request-size cap
// is 4MB vs 20MB for hosted URLs — not enforced here.
func (p *DiagnosisProvider) Analyze(ctx context.Context, imageData []byte) (diagnosis.AnalysisResult, error) {
	mimeType := http.DetectContentType(imageData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imageData))

	body, err := json.Marshal(visionRequest{
		Model: p.model,
		Messages: []visionMessage{
			{
				Role: "user",
				Content: []visionContentPart{
					{Type: "text", Text: diagnosisPrompt},
					{Type: "image_url", ImageURL: &visionImageURL{URL: dataURL}},
				},
			},
		},
		ResponseFormat: &responseFormat{Type: "json_object"},
	})
	if err != nil {
		return diagnosis.AnalysisResult{}, fmt.Errorf("groq: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatCompletionsEndpoint, bytes.NewReader(body))
	if err != nil {
		return diagnosis.AnalysisResult{}, fmt.Errorf("groq: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return diagnosis.AnalysisResult{}, fmt.Errorf("groq: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diagnosis.AnalysisResult{}, fmt.Errorf("groq: status %d", resp.StatusCode)
	}

	var vr visionResponse
	if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil {
		return diagnosis.AnalysisResult{}, fmt.Errorf("groq: decode response: %w", err)
	}
	if len(vr.Choices) == 0 {
		return diagnosis.AnalysisResult{}, fmt.Errorf("groq: no choices returned")
	}

	raw := strings.TrimSpace(vr.Choices[0].Message.Content)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var dj diagnosisJSON
	if err := json.Unmarshal([]byte(raw), &dj); err != nil {
		return diagnosis.AnalysisResult{}, fmt.Errorf("groq: parse diagnosis JSON %q: %w", raw, err)
	}

	return diagnosis.AnalysisResult{Issue: dj.Issue, Cure: dj.Cure, Provider: aiprovider.Groq}, nil
}
