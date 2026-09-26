// Package diagnosis holds the AI Doctor domain: a photo of a plant is
// analyzed and returns an issue and a suggested cure, modeling the
// "Diseases Detection" screen.
package diagnosis

import "time"

// Disclaimer is always attached to a Diagnosis. Per project convention, AI
// Doctor results are assistance, never a guaranteed diagnosis.
const Disclaimer = "This result is AI-assisted and may be inaccurate. " +
	"It is not a substitute for professional horticultural or botanical advice."

// Diagnosis is the result of analyzing a plant photo.
type Diagnosis struct {
	ID         string    `json:"id" bson:"_id"`
	UserID     string    `json:"userId" bson:"userId"`
	PlantID    *string   `json:"plantId,omitempty" bson:"plantId,omitempty"`
	Issue      string    `json:"issue" bson:"issue"`
	Cure       string    `json:"cure" bson:"cure"`
	Disclaimer string    `json:"disclaimer" bson:"disclaimer"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	Provider   string    `json:"provider,omitempty" bson:"provider,omitempty"`
}
