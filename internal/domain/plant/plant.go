// Package plant holds the Plant entity and its repository contract.
// It models the "Maintainance" screens: a user registers a plant and
// receives a generated watering/fertilizer care roadmap, then logs
// waterings/fertilizings and gets reminders when the next one is due.
package plant

import "time"

// CareRoadmap is the generated watering schedule and guidance shown on the
// "Maintainance" screen (watering times, tips, fertilizer suggestion).
type CareRoadmap struct {
	WateringTimes            []string `json:"wateringTimes" bson:"wateringTimes"`
	WaterAmountMl            int      `json:"waterAmountMl" bson:"waterAmountMl"`
	Tips                     string   `json:"tips" bson:"tips"`
	FertilizerRecommendation string   `json:"fertilizerRecommendation" bson:"fertilizerRecommendation"`
	FertilizingIntervalDays  int      `json:"fertilizingIntervalDays" bson:"fertilizingIntervalDays"`
}

// Plant is a plant registered by a user for care tracking.
type Plant struct {
	ID                    string      `json:"id" bson:"_id"`
	UserID                string      `json:"userId" bson:"userId"`
	Name                  string      `json:"name" bson:"name"`
	Type                  string      `json:"type" bson:"type"`
	AgeStage              string      `json:"ageStage" bson:"ageStage"`
	Location              string      `json:"location,omitempty" bson:"location,omitempty"`
	Sunlight              string      `json:"sunlight,omitempty" bson:"sunlight,omitempty"`
	WateringFrequencyDays int         `json:"wateringFrequencyDays,omitempty" bson:"wateringFrequencyDays,omitempty"`
	ImageKey              string      `json:"imageKey,omitempty" bson:"imageKey,omitempty"`
	ImageURL              string      `json:"image,omitempty" bson:"imageUrl,omitempty"`
	CareRoadmap           CareRoadmap `json:"careRoadmap" bson:"careRoadmap"`

	LastWateredAt     *time.Time `json:"lastWateredAt,omitempty" bson:"lastWateredAt,omitempty"`
	NextWateringAt    time.Time  `json:"nextWateringAt" bson:"nextWateringAt"`
	LastFertilizedAt  *time.Time `json:"lastFertilizedAt,omitempty" bson:"lastFertilizedAt,omitempty"`
	NextFertilizingAt *time.Time `json:"nextFertilizingAt,omitempty" bson:"nextFertilizingAt,omitempty"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

// IdentificationResult is what an AI botanist returns when identifying a plant photo.
type IdentificationResult struct {
	Species               string `json:"species"`
	SuggestedNickname     string `json:"suggestedNickname"`
	Location              string `json:"location"`
	Sunlight              string `json:"sunlight"`
	WateringFrequencyDays int    `json:"wateringFrequencyDays"`
	WaterAmountMl         int    `json:"waterAmountMl"`
	Health                int    `json:"health"`
	CareTips              string `json:"careTips"`
}
