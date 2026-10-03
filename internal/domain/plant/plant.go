// Package plant holds the Plant entity and its repository contract.
// It models the "My Plants" screens: a user registers a plant with a
// watering interval, gets tips/fertilizer guidance generated for it, logs
// waterings/fertilizings, and gets reminders when the next one is due.
package plant

import "time"

// CareRoadmap holds generated care guidance for a plant, derived from its
// age stage. FertilizerRecommendation is stored separately on Plant
// (FertilizerNote) since the Flutter client reads it as a top-level field.
type CareRoadmap struct {
	Tips                    string `bson:"tips"`
	FertilizingIntervalDays int    `bson:"fertilizingIntervalDays"`
}

// Plant is a plant registered by a user for care tracking.
//
// JSON field names match the Flutter client's existing model exactly (see
// lib/features/plants/domain/model/plant.dart and
// lib/features/plants/data/models/plant_dto.dart) - Go/BSON names stay
// close to the original backend schema (Name/Type, not Nickname/Species)
// to avoid a data migration and keep MarkWatered/MarkFertilized/Due
// untouched.
type Plant struct {
	ID       string `json:"_id" bson:"_id"`
	UserID   string `json:"userId" bson:"userId"`
	Name     string `json:"nickname" bson:"name"`
	Type     string `json:"species" bson:"type"`
	AgeStage string `json:"ageStage,omitempty" bson:"ageStage,omitempty"`

	Image    string     `json:"image" bson:"image,omitempty"`
	Location string     `json:"location" bson:"location,omitempty"`
	Sunlight string     `json:"sunlight" bson:"sunlight,omitempty"`
	Health   *int       `json:"health,omitempty" bson:"health,omitempty"`
	Status   string     `json:"status" bson:"status,omitempty"`
	Humidity string     `json:"humidity" bson:"humidity,omitempty"`
	LastScan *time.Time `json:"lastScan,omitempty" bson:"lastScan,omitempty"`

	CareRoadmap    CareRoadmap `json:"-" bson:"careRoadmap"`
	FertilizerNote string      `json:"fertilizerNote" bson:"fertilizerNote,omitempty"`

	// WateringFrequencyDays is user-set at creation (the Add Plant
	// screen's "Water every (days)" field) - unlike fertilizing, watering
	// cadence is not derived from the roadmap.
	WateringFrequencyDays int        `json:"wateringFrequencyDays" bson:"wateringFrequencyDays"`
	LastWateredAt         *time.Time `json:"lastWatered,omitempty" bson:"lastWateredAt,omitempty"`
	NextWateringAt        time.Time  `json:"nextWatering" bson:"nextWateringAt"`
	// WaterLevel is a fixed-English display string ("Today", "Tomorrow",
	// "In N days") recomputed alongside NextWateringAt - see
	// usecase/plant/roadmap.go's waterLevelFor for why it's never
	// localized.
	WaterLevel string `json:"waterLevel" bson:"waterLevel"`

	LastFertilizedAt  *time.Time `json:"-" bson:"lastFertilizedAt,omitempty"`
	NextFertilizingAt *time.Time `json:"-" bson:"nextFertilizingAt,omitempty"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}
