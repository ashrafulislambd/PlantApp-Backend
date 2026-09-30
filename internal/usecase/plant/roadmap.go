package plant

import (
	"strings"
	"time"

	"plantpal-backend/internal/domain/plant"
)

// nextWateringTime finds the next clock time in `times` ("HH:MM") after
// `from`, rolling to tomorrow if every time today has passed.
func nextWateringTime(times []string, from time.Time) time.Time {
	var best time.Time
	for _, t := range times {
		parsed, err := time.Parse("15:04", strings.TrimSpace(t))
		if err != nil {
			continue
		}
		c := time.Date(from.Year(), from.Month(), from.Day(), parsed.Hour(), parsed.Minute(), 0, 0, from.Location())
		if !c.After(from) {
			c = c.AddDate(0, 0, 1)
		}
		if best.IsZero() || c.Before(best) {
			best = c
		}
	}
	if best.IsZero() {
		return from.Add(24 * time.Hour)
	}
	return best
}

// buildRoadmap derives a watering schedule, tips, and fertilizer
// suggestion from plant type/age stage — a rule-based placeholder for
// "Create My Roadmap", swappable later for an AI-driven version.
func buildRoadmap(plantType, ageStage, lang string) plant.CareRoadmap {
	bn := lang == "bn"
	waterAmountMl, wateringTimes := 200, []string{"08:00", "18:00"}
	tips := "Don't expose to excess sun; find a cool, dry place with plenty of indirect sunlight."
	fertilizer, fertDays := "Use a balanced NPK fertilizer every 2 weeks.", 14
	if bn {
		tips = "অতিরিক্ত রোদ এড়িয়ে চলুন; পর্যাপ্ত পরোক্ষ আলোযুক্ত একটি ঠান্ডা, শুকনো জায়গা বেছে নিন।"
		fertilizer = "প্রতি ২ সপ্তাহে একটি সুষম NPK সার ব্যবহার করুন।"
	}
	switch strings.ToLower(strings.TrimSpace(ageStage)) {
	case "seed", "seedling":
		waterAmountMl, wateringTimes, fertDays = 100, []string{"08:00"}, 0
		fertilizer = "Avoid fertilizer until the first true leaves appear."
		if bn {
			fertilizer = "প্রথম প্রকৃত পাতা না গজানো পর্যন্ত সার ব্যবহার এড়িয়ে চলুন।"
		}
	case "mature", "adult":
		waterAmountMl, wateringTimes, fertDays = 300, []string{"08:00", "12:00", "18:00"}, 7
		fertilizer = "Use Potassium (K) based fertilizer to support blooming."
		if bn {
			fertilizer = "ফুল ফোটাতে সাহায্য করতে পটাশিয়াম (K) ভিত্তিক সার ব্যবহার করুন।"
		}
	}
	if strings.Contains(strings.ToLower(plantType), "water") {
		waterAmountMl += 100
	}
	return plant.CareRoadmap{
		WateringTimes: wateringTimes, WaterAmountMl: waterAmountMl,
		Tips: tips, FertilizerRecommendation: fertilizer, FertilizingIntervalDays: fertDays,
	}
}
