package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/diagnosis"
)

type DiagnosisRepository struct {
	coll *mongo.Collection
}

func NewDiagnosisRepository(db *mongo.Database) *DiagnosisRepository {
	return &DiagnosisRepository{coll: db.Collection("diagnoses")}
}

// EnsureIndexes supports List, both the plain per-user history and the
// ?plantId= filtered view.
func (r *DiagnosisRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}, {Key: "plantId", Value: 1}, {Key: "createdAt", Value: 1}},
	})
	return err
}

func (r *DiagnosisRepository) Create(ctx context.Context, d *diagnosis.Diagnosis) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.coll.InsertOne(ctx, d)
	if mongo.IsDuplicateKeyError(err) {
		return apperr.ErrConflict
	}
	return err
}

func (r *DiagnosisRepository) GetByID(ctx context.Context, id, userID string) (*diagnosis.Diagnosis, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var d diagnosis.Diagnosis
	err := r.coll.FindOne(ctx, bson.M{"_id": id, "userId": userID}).Decode(&d)
	if err == mongo.ErrNoDocuments {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DiagnosisRepository) List(ctx context.Context, userID string, plantID *string) ([]*diagnosis.Diagnosis, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	if plantID != nil {
		filter["plantId"] = *plantID
	}

	cur, err := r.coll.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := make([]*diagnosis.Diagnosis, 0)
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
