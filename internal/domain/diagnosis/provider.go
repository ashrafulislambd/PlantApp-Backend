package diagnosis

import (
	"context"

	"myplantpal-backend/internal/domain/aiprovider"
)

// AnalysisResult is what an AI provider returns for a submitted photo.
// IssueBn/CureBn are optional Bengali translations; a real AI provider can
// leave them empty and let the caller fall back to English.
type AnalysisResult struct {
	Issue    string
	Cure     string
	IssueBn  string
	CureBn   string
	Provider aiprovider.Name
}

// Provider abstracts the AI plant-diagnosis backend. Per the project's hard
// rule, Flutter never talks to Gemini/Groq directly — only this backend
// does, through an implementation of this interface. The in-memory mock in
// internal/infrastructure/ai satisfies it for now; swap it for a real
// Gemini/Groq-backed implementation later without touching the usecase or
// delivery layers.
type Provider interface {
	Analyze(ctx context.Context, imageData []byte) (AnalysisResult, error)
}
