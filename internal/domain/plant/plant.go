// Package plant holds the Plant entity and its repository contract.
package plant

import "time"

type CareRoadmap struct {
	WateringTimes            []string `json:"wateringTimes"`
	WaterAmountMl            int      `json:"waterAmountMl"`
	Tips                     string   `json:"tips"`
	FertilizerRecommendation string   `json:"fertilizerRecommendation"`
	FertilizingIntervalDays  int      `json:"fertilizingIntervalDays"`
}

type Plant struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	AgeStage    string      `json:"ageStage"`
	CareRoadmap CareRoadmap `json:"careRoadmap"`

	LastWateredAt     *time.Time `json:"lastWateredAt,omitempty"`
	NextWateringAt    time.Time  `json:"nextWateringAt"`
	LastFertilizedAt  *time.Time `json:"lastFertilizedAt,omitempty"`
	NextFertilizingAt *time.Time `json:"nextFertilizingAt,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}