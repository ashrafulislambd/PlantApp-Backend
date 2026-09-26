package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/chat"
)

type ChatRepository struct {
	coll *mongo.Collection
}

func NewChatRepository(db *mongo.Database) *ChatRepository {
	return &ChatRepository{coll: db.Collection("chat_messages")}
}

// EnsureIndexes supports ListBySession, which always filters by userId +
// sessionId and sorts by createdAt.
func (r *ChatRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}, {Key: "sessionId", Value: 1}, {Key: "createdAt", Value: 1}},
	})
	return err
}

func (r *ChatRepository) Create(ctx context.Context, m *chat.Message) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.coll.InsertOne(ctx, m)
	if mongo.IsDuplicateKeyError(err) {
		return apperr.ErrConflict
	}
	return err
}

func (r *ChatRepository) ListBySession(ctx context.Context, userID, sessionID string) ([]*chat.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID, "sessionId": sessionID}
	cur, err := r.coll.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := make([]*chat.Message, 0)
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
