package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"plantpal-backend/internal/domain/plant"
)

type PlantEventRepository struct {
	coll *mongo.Collection
}

func NewPlantEventRepository(db *mongo.Database) *PlantEventRepository {
	return &PlantEventRepository{coll: db.Collection("plant_events")}
}

func (r *PlantEventRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "plantId", Value: 1}, {Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}}},
	})
	return err
}

func (r *PlantEventRepository) Create(ctx context.Context, event *plant.PlantEvent) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.coll.InsertOne(ctx, event)
	return err
}

func (r *PlantEventRepository) ListByPlant(ctx context.Context, plantID, userID string) ([]*plant.PlantEvent, error) {
	return r.list(ctx, bson.M{"plantId": plantID, "userId": userID})
}

func (r *PlantEventRepository) ListByUser(ctx context.Context, userID string) ([]*plant.PlantEvent, error) {
	return r.list(ctx, bson.M{"userId": userID})
}

func (r *PlantEventRepository) list(ctx context.Context, filter bson.M) ([]*plant.PlantEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cur, err := r.coll.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := make([]*plant.PlantEvent, 0)
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
