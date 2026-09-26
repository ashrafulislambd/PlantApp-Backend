// Package plant holds the Plant entity and its repository contract.
// It models the "Maintainance" screens: a user registers a plant and
// receives a generated watering/fertilizer care roadmap.
package plant

import "time"

// CareRoadmap is the generated watering schedule and guidance shown on the
// "Maintainance" screen (watering times, tips, fertilizer suggestion).
type CareRoadmap struct {
	WateringTimes            []string `json:"wateringTimes" bson:"wateringTimes"`
	WaterAmountMl            int      `json:"waterAmountMl" bson:"waterAmountMl"`
	Tips                     string   `json:"tips" bson:"tips"`
	FertilizerRecommendation string   `json:"fertilizerRecommendation" bson:"fertilizerRecommendation"`
}

// Plant is a plant registered by a user for care tracking.
type Plant struct {
	ID          string      `json:"id" bson:"_id"`
	UserID      string      `json:"userId" bson:"userId"`
	Name        string      `json:"name" bson:"name"`
	Type        string      `json:"type" bson:"type"`
	AgeStage    string      `json:"ageStage" bson:"ageStage"`
	CareRoadmap CareRoadmap `json:"careRoadmap" bson:"careRoadmap"`
	CreatedAt   time.Time   `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt" bson:"updatedAt"`
}
