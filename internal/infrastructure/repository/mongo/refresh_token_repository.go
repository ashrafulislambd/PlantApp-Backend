package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/refreshtoken"
)

type RefreshTokenRepository struct {
	coll *mongo.Collection
}

func NewRefreshTokenRepository(db *mongo.Database) *RefreshTokenRepository {
	return &RefreshTokenRepository{coll: db.Collection("refresh_tokens")}
}

func (r *RefreshTokenRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "tokenHash", Value: 1}}})
	return err
}

func (r *RefreshTokenRepository) Create(ctx context.Context, t *refreshtoken.RefreshToken) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.coll.InsertOne(ctx, t)
	return err
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*refreshtoken.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var t refreshtoken.RefreshToken
	err := r.coll.FindOne(ctx, bson.M{"tokenHash": hash}).Decode(&t)
	if err == mongo.ErrNoDocuments {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	now := time.Now().UTC()
	_, err := r.coll.UpdateByID(ctx, id, bson.M{"$set": bson.M{"revokedAt": now}})
	return err
}
