package diagnosis

import "time"

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
