package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/order"
)

type OrderRepository struct {
	coll *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) *OrderRepository {
	return &OrderRepository{coll: db.Collection("orders")}
}

// EnsureIndexes supports List (a user's order history, newest first).
func (r *OrderRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}},
	})
	return err
}

func (r *OrderRepository) Create(ctx context.Context, o *order.Order) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.coll.InsertOne(ctx, o)
	if mongo.IsDuplicateKeyError(err) {
		return apperr.ErrConflict
	}
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id, userID string) (*order.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var o order.Order
	err := r.coll.FindOne(ctx, bson.M{"_id": id, "userId": userID}).Decode(&o)
	if err == mongo.ErrNoDocuments {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) List(ctx context.Context, userID string) ([]*order.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cur, err := r.coll.Find(ctx, bson.M{"userId": userID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := make([]*order.Order, 0)
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
