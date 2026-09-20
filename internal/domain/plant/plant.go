// Package plant holds the Plant entity and its repository contract.
// It models the "Maintainance" screens: a user registers a plant and
// receives a generated watering/fertilizer care roadmap.
package plant

import "time"

// CareRoadmap is the generated watering schedule and guidance shown on the
// "Maintainance" screen (watering times, tips, fertilizer suggestion).
type CareRoadmap struct {
	WateringTimes            []string `json:"wateringTimes"`
	WaterAmountMl            int      `json:"waterAmountMl"`
	Tips                     string   `json:"tips"`
	FertilizerRecommendation string   `json:"fertilizerRecommendation"`
}

// Plant is a plant registered by a user for care tracking.
type Plant struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	AgeStage    string      `json:"ageStage"`
	CareRoadmap CareRoadmap `json:"careRoadmap"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}
