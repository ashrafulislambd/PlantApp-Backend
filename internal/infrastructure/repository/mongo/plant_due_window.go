package mongo

import (
"context"
"time"

"go.mongodb.org/mongo-driver/v2/bson"
"go.mongodb.org/mongo-driver/v2/mongo"

"plantpal-backend/internal/domain/plant"
)

// EnsureDueIndexes speeds up the reminder scheduler's scan across all users.
func (r *PlantRepository) EnsureDueIndexes(ctx context.Context) error {
_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
{Keys: bson.D{{Key: "nextWateringAt", Value: 1}}},
{Keys: bson.D{{Key: "nextFertilizingAt", Value: 1}}},
})
return err
}

// ListDueBetween returns plants of every user whose next watering or
// fertilizing falls in (after, before]. Used by the reminder scheduler.
func (r *PlantRepository) ListDueBetween(ctx context.Context, after, before time.Time) ([]*plant.Plant, error) {
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()
window := bson.M{"$gt": after, "$lte": before}
filter := bson.M{
"$or": bson.A{
bson.M{"nextWateringAt": window},
bson.M{"nextFertilizingAt": window},
},
}
cur, err := r.coll.Find(ctx, filter)
if err != nil {
return nil, err
}
defer cur.Close(ctx)

out := make([]*plant.Plant, 0)
if err := cur.All(ctx, &out); err != nil {
return nil, err
}
return out, nil
}