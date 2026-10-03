package plant

import (
	"fmt"
	"strings"
	"time"

	"plantpal-backend/internal/domain/plant"
)

// buildRoadmap derives tips and a fertilizer suggestion from age stage - a
// rule-based placeholder for "Create My Roadmap", swappable later for an
// AI-driven version. Returns the roadmap (tips + fertilizing interval) and
// the fertilizer recommendation text separately, since the latter is
// stored on Plant.FertilizerNote (a top-level field the Flutter client
// reads directly).
func buildRoadmap(ageStage, lang string) (plant.CareRoadmap, string) {
	bn := lang == "bn"
	tips := "Don't expose to excess sun; find a cool, dry place with plenty of indirect sunlight."
	fertilizer, fertDays := "Use a balanced NPK fertilizer every 2 weeks.", 14
	if bn {
		tips = "অতিরিক্ত রোদ এড়িয়ে চলুন; পর্যাপ্ত পরোক্ষ আলোযুক্ত একটি ঠান্ডা, শুকনো জায়গা বেছে নিন।"
		fertilizer = "প্রতি ২ সপ্তাহে একটি সুষম NPK সার ব্যবহার করুন।"
	}
	switch strings.ToLower(strings.TrimSpace(ageStage)) {
	case "seed", "seedling":
		fertDays = 0
		fertilizer = "Avoid fertilizer until the first true leaves appear."
		if bn {
			fertilizer = "প্রথম প্রকৃত পাতা না গজানো পর্যন্ত সার ব্যবহার এড়িয়ে চলুন।"
		}
	case "mature", "adult":
		fertDays = 7
		fertilizer = "Use Potassium (K) based fertilizer to support blooming."
		if bn {
			fertilizer = "ফুল ফোটাতে সাহায্য করতে পটাশিয়াম (K) ভিত্তিক সার ব্যবহার করুন।"
		}
	}
	return plant.CareRoadmap{Tips: tips, FertilizingIntervalDays: fertDays}, fertilizer
}

// recomputeWatering sets NextWateringAt (LastWateredAt, or now if never
// watered, plus WateringFrequencyDays) and the display WaterLevel. Called
// on create, on any update that touches LastWateredAt or
// WateringFrequencyDays, and by MarkWatered.
func recomputeWatering(p *plant.Plant, now time.Time) {
	base := now
	if p.LastWateredAt != nil {
		base = *p.LastWateredAt
	}
	days := p.WateringFrequencyDays
	if days < 1 {
		days = 7
	}
	p.NextWateringAt = base.AddDate(0, 0, days)
	p.WaterLevel = waterLevelFor(p.NextWateringAt, now)
}

// waterLevelFor is a fixed-English display string - the Flutter client
// checks `waterLevel == 'Today'` literally (see
// lib/features/plants/presentation/providers/plants_provider.dart's
// waterTodayCount), so this must not be localized.
func waterLevelFor(next, now time.Time) string {
	days := int(next.Truncate(24*time.Hour).Sub(now.Truncate(24*time.Hour)).Hours() / 24)
	switch {
	case days <= 0:
		return "Today"
	case days == 1:
		return "Tomorrow"
	default:
		return fmt.Sprintf("In %d days", days)
	}
}
