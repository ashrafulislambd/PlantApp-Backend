package diagnosis

import "context"

// AnalysisResult is what an AI provider returns for a submitted photo.
type AnalysisResult struct {
	Issue string
	Cure  string
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
