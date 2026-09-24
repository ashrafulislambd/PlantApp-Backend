// Package fertilizer holds the Fertilizer entity and its repository
// contract, modeling the "Fertilizer Making" screen: a searchable list of
// homemade fertilizer recipes.
package fertilizer

import "time"

// Fertilizer is a homemade fertilizer recipe, e.g. "Nitrogen (leaf growth)".
//
// The Bn fields hold Bengali translations for seed data; they're excluded
// from JSON directly (json:"-") and only surfaced through Localized, which
// picks the requested language and falls back to the base (English)
// fields when no translation exists — e.g. for fertilizers a user adds
// themselves, which only exist in whatever language they typed.
type Fertilizer struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	Instructions string    `json:"instructions"`
	CreatedAt    time.Time `json:"createdAt"`

	NameBn         string `json:"-"`
	CategoryBn     string `json:"-"`
	InstructionsBn string `json:"-"`
}

// Localized returns a copy with Name/Category/Instructions swapped for
// their Bengali translation when lang is "bn" and a translation exists.
func (f Fertilizer) Localized(lang string) Fertilizer {
	if lang != "bn" || f.NameBn == "" {
		return f
	}
	f.Name = f.NameBn
	f.Category = f.CategoryBn
	f.Instructions = f.InstructionsBn
	return f
}
