package fertilizer

import "time"

// Fertilizer is a homemade organic fertilizer recipe shown on the
// Fertilizer screen.
type Fertilizer struct {
	ID           string    `json:"id" bson:"_id"`
	Name         string    `json:"name" bson:"name"`
	Category     string    `json:"category" bson:"category"`
	Instructions string    `json:"instructions" bson:"instructions"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`

	NameBn         string `json:"-" bson:"nameBn,omitempty"`
	CategoryBn     string `json:"-" bson:"categoryBn,omitempty"`
	InstructionsBn string `json:"-" bson:"instructionsBn,omitempty"`
}

// Localized returns a copy with Name/Category/Instructions swapped for
// their Bengali translation when lang is "bn" and a translation exists.
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
