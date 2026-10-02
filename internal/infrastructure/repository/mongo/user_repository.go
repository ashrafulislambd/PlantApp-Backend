package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/user"
)

type UserRepository struct {
	coll *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{coll: db.Collection("users")}
}

func (r *UserRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			// Unique per Google account, but only for users that have one
			// (email/password users have no googleId field at all).
			Keys: bson.D{{Key: "googleId", Value: 1}},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(
				bson.D{{Key: "googleId", Value: bson.D{{Key: "$type", Value: "string"}}}},
			),
		},
	})
	return err
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.coll.InsertOne(ctx, u)
	if mongo.IsDuplicateKeyError(err) {
		return apperr.ErrConflict
	}
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var u user.User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*user.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var u user.User
	err := r.coll.FindOne(ctx, bson.M{"googleId": googleID}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) LinkGoogle(ctx context.Context, id, googleID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"googleId": googleID, "updatedAt": time.Now().UTC()}},
	)
	if mongo.IsDuplicateKeyError(err) {
		return apperr.ErrConflict
	}
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdatePasswordHash(ctx context.Context, id, passwordHash string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"passwordHash": passwordHash, "updatedAt": time.Now().UTC()}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var u user.User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
