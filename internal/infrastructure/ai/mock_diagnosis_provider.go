// Package ai holds mock implementations of the domain AI provider
// interfaces (diagnosis.Provider, chat.ReplyProvider). Swap these for
// Gemini/Groq-backed implementations later — the usecase layer only
// depends on the interfaces, not on this package.
package ai

import (
	"context"
	"sync/atomic"

	"myplantpal-backend/internal/domain/diagnosis"
)

var mockDiagnoses = []diagnosis.AnalysisResult{
	{
		Issue:   "Your plant is suffering from Phosphorus deficiency.",
		Cure:    "Use Triple Superphosphate (TSP) or Single Superphosphate (SSP).",
		IssueBn: "আপনার গাছটি ফসফরাসের অভাবে ভুগছে।",
		CureBn:  "ট্রিপল সুপারফসফেট (TSP) অথবা সিঙ্গেল সুপারফসফেট (SSP) ব্যবহার করুন।",
	},
	{
		Issue:   "Your plant shows early signs of Nitrogen deficiency (pale, yellowing leaves).",
		Cure:    "Apply a nitrogen-rich homemade fertilizer, such as a banana peel and tea leaf soak.",
		IssueBn: "আপনার গাছে নাইট্রোজেন ঘাটতির প্রাথমিক লক্ষণ দেখা যাচ্ছে (ফ্যাকাশে, হলুদ পাতা)।",
		CureBn:  "কলার খোসা ও চা পাতা ভেজানো পানির মতো নাইট্রোজেন সমৃদ্ধ ঘরোয়া সার প্রয়োগ করুন।",
	},
	{
		Issue:   "Your plant's leaves show signs of Potassium deficiency (brown leaf edges).",
		Cure:    "Work dried, crushed banana peels into the soil near the roots.",
		IssueBn: "আপনার গাছের পাতায় পটাশিয়াম ঘাটতির লক্ষণ দেখা যাচ্ছে (পাতার কিনারা বাদামী)।",
		CureBn:  "শুকনো, গুঁড়ো করা কলার খোসা গাছের মূলের কাছের মাটিতে মিশিয়ে দিন।",
	},
}

// MockDiagnosisProvider cycles through a small set of canned results so
// the API is exercisable end to end without a real AI integration.
type MockDiagnosisProvider struct {
	counter atomic.Uint64
}

func NewMockDiagnosisProvider() *MockDiagnosisProvider {
	return &MockDiagnosisProvider{}
}

func (p *MockDiagnosisProvider) Analyze(_ context.Context, _ []byte) (diagnosis.AnalysisResult, error) {
	i := p.counter.Add(1) - 1
	return mockDiagnoses[int(i)%len(mockDiagnoses)], nil
}
