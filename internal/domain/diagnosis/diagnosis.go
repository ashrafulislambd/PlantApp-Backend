package diagnosis

import (
	"fmt"
	"strings"
	"time"
)

// Disclaimer is always attached to a Diagnosis. Per project convention, AI
// Doctor results are assistance, never a guaranteed diagnosis.
const Disclaimer = "This result is AI-assisted and may be inaccurate. " +
	"It is not a substitute for professional horticultural or botanical advice."

const DisclaimerBn = "এই তথ্যটি AI-এর সাহায্যে তৈরি, তাই এতে ভুল থাকতে পারে। " +
	"সঠিক সিদ্ধান্তের জন্য অনুগ্রহ করে উদ্ভিদ বিশেষজ্ঞের পরামর্শ নিন।"

// Diagnosis is the result of analyzing a plant photo.
//
// The Bn fields hold Bengali translations of the result; they're excluded
// from JSON directly (json:"-") and only surfaced through Localized. They
// still need bson tags (also "-"-style via a dedicated key) so they persist
// to Mongo alongside the rest of the document.
type Diagnosis struct {
	ID         string    `json:"id" bson:"_id"`
	UserID     string    `json:"userId" bson:"userId"`
	PlantID    *string   `json:"plantId,omitempty" bson:"plantId,omitempty"`
	Issue      string    `json:"issue" bson:"issue"`
	Cure       string    `json:"cure" bson:"cure"`
	Confidence string    `json:"confidence,omitempty" bson:"confidence,omitempty"`
	Severity   string    `json:"severity,omitempty" bson:"severity,omitempty"`
	Fertilizer string    `json:"fertilizer,omitempty" bson:"fertilizer,omitempty"`
	Disclaimer string    `json:"disclaimer" bson:"disclaimer"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	Provider   string    `json:"provider,omitempty" bson:"provider,omitempty"`

	IssueBn      string `json:"-" bson:"issueBn,omitempty"`
	CureBn       string `json:"-" bson:"cureBn,omitempty"`
	DisclaimerBn string `json:"-" bson:"disclaimerBn,omitempty"`

ImageKey         string `json:"-" bson:"imageKey,omitempty"`
ImageContentType string `json:"-" bson:"imageContentType,omitempty"`
ImageURL         string `json:"imageUrl,omitempty" bson:"-"`
}

// Localized returns a copy with Issue/Cure/Disclaimer swapped for their
// Bengali translation when lang is "bn" and a translation exists.
func (d Diagnosis) Localized(lang string) Diagnosis {
	if lang != "bn" || d.IssueBn == "" {
		return d
	}
	d.Issue = d.IssueBn
	d.Cure = d.CureBn
	d.Disclaimer = d.DisclaimerBn
	return d
}

// PromptText renders the scan as one bracketed line a language model can
// read, e.g. "[Scan result: issue ...; cure ...]". Always English, whatever
// the user's language, so the model sees consistent field names.
func (d Diagnosis) PromptText() string {
	parts := []string{"issue " + d.Issue, "cure " + d.Cure}
	if d.Severity != "" {
		parts = append(parts, "severity "+d.Severity)
	}
	if d.Confidence != "" {
		parts = append(parts, "confidence "+d.Confidence)
	}
	if d.Fertilizer != "" {
		parts = append(parts, "fertilizer "+d.Fertilizer)
	}
	return "[Scan result: " + strings.Join(parts, "; ") + "]"
}

// ChatSummary is the short assistant message saved in a chat session when a
// photo is scanned from the chat, in the user's language (Bengali when lang
// is "bn" and a translation exists). Markdown, like every assistant reply.
func (d Diagnosis) ChatSummary(lang string) string {
	l := d.Localized(lang)
	if lang == "bn" && d.IssueBn != "" {
		return fmt.Sprintf("আপনার ছবিটি বিশ্লেষণ করেছি।\n\n**সমস্যা:** %s\n\n**প্রস্তাবিত চিকিৎসা:** %s", l.Issue, l.Cure)
	}
	return fmt.Sprintf("I analysed your photo.\n\n**Issue:** %s\n\n**Suggested treatment:** %s", l.Issue, l.Cure)
}
