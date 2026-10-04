package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/chat"
)

type ChatRepository struct {
	coll *mongo.Collection
}

func NewChatRepository(db *mongo.Database) *ChatRepository {
	return &ChatRepository{coll: db.Collection("chat_messages")}
}

// EnsureIndexes supports ListBySession (filters by userId + sessionId and
// sorts by createdAt) and ListSessions (filters by userId and sorts by
// createdAt before grouping).
func (r *ChatRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "sessionId", Value: 1}, {Key: "createdAt", Value: 1}}},
		{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: 1}}},
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

// maxSessions caps how many conversations the history menu can list.
const maxSessions = 200

// sessionRow is one $group output row of ListSessions.
type sessionRow struct {
	ID           string    `bson:"_id"`
	FirstContent string    `bson:"firstContent"`
	FirstDiagID  string    `bson:"firstDiagnosisId"`
	LastAt       time.Time `bson:"lastAt"`
	Count        int       `bson:"count"`
}

// ListSessions groups the user's messages by session. Messages are sorted
// oldest-first before grouping so $first yields the opening message, which
// is always the user's (Send stores it before asking the AI) and becomes the
// session title.
func (r *ChatRepository) ListSessions(ctx context.Context, userID string) ([]*chat.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "userId", Value: userID}}}},
		{{Key: "$sort", Value: bson.D{{Key: "createdAt", Value: 1}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$sessionId"},
			{Key: "firstContent", Value: bson.D{{Key: "$first", Value: "$content"}}},
			{Key: "firstDiagnosisId", Value: bson.D{{Key: "$first", Value: "$diagnosisId"}}},
			{Key: "lastAt", Value: bson.D{{Key: "$max", Value: "$createdAt"}}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "lastAt", Value: -1}}}},
		{{Key: "$limit", Value: maxSessions}},
	}
	cur, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var rows []sessionRow
	if err := cur.All(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]*chat.Session, 0, len(rows))
	for _, row := range rows {
		out = append(out, &chat.Session{
			ID:            row.ID,
			Title:         chat.SessionTitle(row.FirstContent, row.FirstDiagID),
			LastMessageAt: row.LastAt,
			MessageCount:  row.Count,
		})
	}
	return out, nil
}

func (r *ChatRepository) DeleteSession(ctx context.Context, userID, sessionID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := r.coll.DeleteMany(ctx, bson.M{"userId": userID, "sessionId": sessionID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return apperr.ErrNotFound
	}
	return nil
}
