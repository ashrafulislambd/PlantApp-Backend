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

	NameBn         string `json:"-" bson:"nameBn,omitempty"`
	CategoryBn     string `json:"-" bson:"categoryBn,omitempty"`
	InstructionsBn string `json:"-" bson:"instructionsBn,omitempty"`
}

// Localized returns a copy with Bengali text when translations are available.
func (f Fertilizer) Localized(lang string) Fertilizer {
	if lang != "bn" || f.NameBn == "" {
		return f
	}
	f.Name = f.NameBn
	if f.CategoryBn != "" {
		f.Category = f.CategoryBn
	}
	if f.InstructionsBn != "" {
		f.Instructions = f.InstructionsBn
	}
	return f
}
