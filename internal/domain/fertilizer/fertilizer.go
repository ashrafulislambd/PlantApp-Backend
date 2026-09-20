// Package fertilizer holds the Fertilizer entity and its repository
// contract, modeling the "Fertilizer Making" screen: a searchable list of
// homemade fertilizer recipes.
package fertilizer

import "time"

// Fertilizer is a homemade fertilizer recipe, e.g. "Nitrogen (leaf growth)".
type Fertilizer struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	Instructions string    `json:"instructions"`
	CreatedAt    time.Time `json:"createdAt"`
}
