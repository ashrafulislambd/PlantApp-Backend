// Package diagnosis holds the AI Doctor domain: a photo of a plant is
// analyzed and returns an issue and a suggested cure, modeling the
// "Diseases Detection" screen.
package diagnosis

import "time"

// Disclaimer is always attached to a Diagnosis. Per project convention, AI
// Doctor results are assistance, never a guaranteed diagnosis.
const Disclaimer = "This result is AI-assisted and may be inaccurate. " +
	"It is not a substitute for professional horticultural or botanical advice."

const DisclaimerBn = "এই ফলাফল AI-সহায়তায় প্রাপ্ত এবং ভুল হতে পারে। " +
	"এটি পেশাদার উদ্ভিদবিদ্যা পরামর্শের বিকল্প নয়।"

// Diagnosis is the result of analyzing a plant photo.
//
// The Bn fields hold Bengali translations of the mocked result; they're
// excluded from JSON directly (json:"-") and only surfaced through
// Localized.
type Diagnosis struct {
	ID         string    `json:"id"`
	PlantID    *string   `json:"plantId,omitempty"`
	Issue      string    `json:"issue"`
	Cure       string    `json:"cure"`
	Disclaimer string    `json:"disclaimer"`
	CreatedAt  time.Time `json:"createdAt"`
	Provider   string    `json:"provider,omitempty"`

	IssueBn      string `json:"-"`
	CureBn       string `json:"-"`
	DisclaimerBn string `json:"-"`
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