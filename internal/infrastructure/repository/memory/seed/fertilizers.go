// Package seed provides initial in-memory data so the API is useful
// out of the box, without a real database behind it yet.
package seed

import (
	"time"

	"myplantpal-backend/internal/domain/fertilizer"
)

// Fertilizers returns the starter recipes shown in the Figma design plus a
// couple of siblings covering the other two primary nutrients.
func Fertilizers() []*fertilizer.Fertilizer {
	now := time.Now().UTC()
	return []*fertilizer.Fertilizer{
		{
			ID:       "fert_seed_nitrogen",
			Name:     "Nitrogen (leaf growth)",
			Category: "Nitrogen",
			Instructions: "Put 1 banana peel + used tea leaves + 1 liter water in a bottle/jar. " +
				"Keep it in a shady, cool place (not under direct sun).\n" +
				"Soak for 2 days.\n" +
				"Then pour a little around the soil near the roots, not directly on the trunk.\n" +
				"Use it once every 10-15 days.",
			CreatedAt: now,
		},
		{
			ID:       "fert_seed_phosphorus",
			Name:     "Phosphorus (root & flower growth)",
			Category: "Phosphorus",
			Instructions: "Crush dried eggshells into a fine powder and mix into the top layer of soil. " +
				"Alternatively, soak rice water for 24 hours and use it to water the plant once a week.",
			CreatedAt: now,
		},
		{
			ID:       "fert_seed_potassium",
			Name:     "Potassium (blooming & disease resistance)",
			Category: "Potassium",
			Instructions: "Save banana peels and dry them in the sun for a few days, then bury the dried " +
				"peels a few centimeters deep near the plant's roots. Water lightly afterward.",
			CreatedAt: now,
		},
	}
}
