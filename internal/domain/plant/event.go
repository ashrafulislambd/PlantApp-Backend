package plant

import "time"

const (
	EventWater     = "water"
	EventFertilize = "fertilize"
	EventScan      = "scan"
	EventSkip      = "skip"
	EventNote      = "note"
)

// PlantEvent is an immutable record of an action or note in a user's garden.
type PlantEvent struct {
	ID        string         `json:"id" bson:"_id"`
	PlantID   string         `json:"plantId" bson:"plantId"`
	UserID    string         `json:"userId" bson:"userId"`
	EventType string         `json:"eventType" bson:"eventType"`
	Reason    string         `json:"reason,omitempty" bson:"reason,omitempty"`
	Note      string         `json:"note,omitempty" bson:"note,omitempty"`
	Points    int            `json:"points" bson:"points"`
	ImageURL  string         `json:"imageUrl,omitempty" bson:"imageUrl,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt time.Time      `json:"createdAt" bson:"createdAt"`
}
