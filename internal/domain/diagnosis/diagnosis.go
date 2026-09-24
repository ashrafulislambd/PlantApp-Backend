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
	ID         string    `json:"id"`
	PlantID    *string   `json:"plantId,omitempty"`
	Issue      string    `json:"issue"`
	Cure       string    `json:"cure"`
	Disclaimer string    `json:"disclaimer"`
	CreatedAt  time.Time `json:"createdAt"`
	Provider   string    `json:"provider,omitempty"`
}
