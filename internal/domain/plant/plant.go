// Package plant holds the Plant entity and its repository contract.
// It models the "My Plants" screens: a user registers a plant with a
// watering interval, gets tips/fertilizer guidance generated for it, logs
// waterings/fertilizings, and gets reminders when the next one is due.
package plant

import (
	"encoding/json"
	"time"
)

// CareRoadmap holds generated care guidance for a plant, derived from its
// age stage and type.
type CareRoadmap struct {
	WateringTimes            []string `json:"wateringTimes,omitempty" bson:"wateringTimes,omitempty"`
	WaterAmountMl            int      `json:"waterAmountMl,omitempty" bson:"waterAmountMl,omitempty"`
	Tips                     string   `json:"tips" bson:"tips"`
	FertilizerRecommendation string   `json:"fertilizerRecommendation,omitempty" bson:"fertilizerRecommendation,omitempty"`
	FertilizingIntervalDays  int      `json:"fertilizingIntervalDays" bson:"fertilizingIntervalDays"`
}

// Plant is a plant registered by a user for care tracking.
type Plant struct {
	ID                    string      `json:"id" bson:"_id"`
	UserID                string      `json:"userId" bson:"userId"`
	Name                  string      `json:"name" bson:"name"`
	Type                  string      `json:"type" bson:"type"`
	AgeStage              string      `json:"ageStage,omitempty" bson:"ageStage,omitempty"`
	Location              string      `json:"location,omitempty" bson:"location,omitempty"`
	Sunlight              string      `json:"sunlight,omitempty" bson:"sunlight,omitempty"`
	// Outdoor is true for plants that live outside (garden, balcony): they get
	// the weather-aware tips (e.g. "raining: skip watering"). Default: indoor.
	Outdoor               bool        `json:"outdoor" bson:"outdoor"`
	WateringFrequencyDays int         `json:"wateringFrequencyDays,omitempty" bson:"wateringFrequencyDays,omitempty"`
	ImageKey              string      `json:"imageKey,omitempty" bson:"imageKey,omitempty"`
	ImageURL              string      `json:"image,omitempty" bson:"imageUrl,omitempty"`
	CareRoadmap           CareRoadmap `json:"careRoadmap" bson:"careRoadmap"`
	FertilizerNote        string      `json:"fertilizerNote,omitempty" bson:"fertilizerNote,omitempty"`

	Status                string      `json:"status,omitempty" bson:"status,omitempty"`
	Humidity              string      `json:"humidity,omitempty" bson:"humidity,omitempty"`

	LastWateredAt         *time.Time  `json:"lastWateredAt,omitempty" bson:"lastWateredAt,omitempty"`
	NextWateringAt        time.Time   `json:"nextWateringAt" bson:"nextWateringAt"`
	WaterLevel            string      `json:"waterLevel,omitempty" bson:"waterLevel,omitempty"`

	LastFertilizedAt      *time.Time  `json:"lastFertilizedAt,omitempty" bson:"lastFertilizedAt,omitempty"`
	NextFertilizingAt     *time.Time  `json:"nextFertilizingAt,omitempty" bson:"nextFertilizingAt,omitempty"`

	CreatedAt             time.Time   `json:"createdAt" bson:"createdAt"`
	UpdatedAt             time.Time   `json:"updatedAt" bson:"updatedAt"`

	// Computed on every read by the plant service (see health.go); never stored.
	Health                int            `json:"health" bson:"-"`
	HealthState           string         `json:"healthState" bson:"-"`
	HealthReasons         []HealthReason `json:"healthReasons" bson:"-"`
	LastScan              *LastScan      `json:"lastScan,omitempty" bson:"-"`
}

// MarshalJSON includes field aliases expected by Flutter client models:
// _id, nickname, species, lastWatered, nextWatering.
func (p *Plant) MarshalJSON() ([]byte, error) {
	type Alias Plant
	return json.Marshal(&struct {
		*Alias
		AltID           string     `json:"_id"`
		AltNickname     string     `json:"nickname"`
		AltSpecies      string     `json:"species"`
		AltLastWatered  *time.Time `json:"lastWatered,omitempty"`
		AltNextWatering time.Time  `json:"nextWatering"`
	}{
		Alias:           (*Alias)(p),
		AltID:           p.ID,
		AltNickname:     p.Name,
		AltSpecies:      p.Type,
		AltLastWatered:  p.LastWateredAt,
		AltNextWatering: p.NextWateringAt,
	})
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
