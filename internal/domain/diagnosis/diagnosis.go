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
	Disclaimer string    `json:"disclaimer" bson:"disclaimer"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	Provider   string    `json:"provider,omitempty" bson:"provider,omitempty"`

	IssueBn      string `json:"-" bson:"issueBn,omitempty"`
	CureBn       string `json:"-" bson:"cureBn,omitempty"`
	DisclaimerBn string `json:"-" bson:"disclaimerBn,omitempty"`
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
