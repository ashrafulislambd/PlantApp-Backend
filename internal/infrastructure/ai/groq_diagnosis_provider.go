package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"myplantpal-backend/internal/domain/diagnosis"
)

const diagnosisSystemPrompt = `You are a plant-health assistant analyzing a photo of a plant. Identify the most
likely issue (nutrient deficiency, pest, disease, watering problem, or say the plant looks
healthy) and a practical home cure. This is assistance, not a guaranteed diagnosis — keep
wording accordingly (e.g. "looks like", "likely"), never absolute claims.

Respond with ONLY a JSON object in this exact shape, with the Bengali fields as a natural
translation of the English ones:
{"issue": "...", "cure": "...", "issueBn": "...", "cureBn": "..."}`

// GroqDiagnosisProvider analyzes plant photos via Groq's vision-capable
// chat completions API.
type GroqDiagnosisProvider struct {
	client *groqClient
	model  string
}

func NewGroqDiagnosisProvider(apiKey, model string) *GroqDiagnosisProvider {
	return &GroqDiagnosisProvider{client: newGroqClient(apiKey), model: model}
}

func (p *GroqDiagnosisProvider) Analyze(ctx context.Context, imageData []byte) (diagnosis.AnalysisResult, error) {
	mimeType := http.DetectContentType(imageData)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imageData))

	reply, err := p.client.complete(ctx, groqChatRequest{
		Model: p.model,
		Messages: []groqMessage{
			{
				Role: "user",
				Content: []any{
					groqTextContent{Type: "text", Text: diagnosisSystemPrompt},
					groqImageContent{Type: "image_url", ImageURL: groqImageURL{URL: dataURI}},
				},
			},
		},
		Temperature:    0.4,
		ResponseFormat: &groqResponseFormat{Type: "json_object"},
	})
	if err != nil {
		return diagnosis.AnalysisResult{}, fmt.Errorf("groq diagnosis: %w", err)
	}

	var parsed struct {
		Issue   string `json:"issue"`
		Cure    string `json:"cure"`
		IssueBn string `json:"issueBn"`
		CureBn  string `json:"cureBn"`
	}
	if err := json.Unmarshal([]byte(reply), &parsed); err != nil {
		return diagnosis.AnalysisResult{}, fmt.Errorf("groq diagnosis: decode model output: %w", err)
	}

	return diagnosis.AnalysisResult{
		Issue:   parsed.Issue,
		Cure:    parsed.Cure,
		IssueBn: parsed.IssueBn,
		CureBn:  parsed.CureBn,
	}, nil
}
