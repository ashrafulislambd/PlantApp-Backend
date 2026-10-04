package mongo

import (
"context"
"time"

"go.mongodb.org/mongo-driver/v2/bson"
"go.mongodb.org/mongo-driver/v2/mongo"
"go.mongodb.org/mongo-driver/v2/mongo/options"

"plantpal-backend/internal/domain/notification"
)

const sentLogTTL = 30 * 24 * 60 * 60 // seconds

type NotificationRepository struct {
devices *mongo.Collection
sent    *mongo.Collection
}

func NewNotificationRepository(db *mongo.Database) *NotificationRepository {
return &NotificationRepository{
devices: db.Collection("device_tokens"),
sent:    db.Collection("reminder_log"),
}
}

func (r *NotificationRepository) EnsureIndexes(ctx context.Context) error {
if _, err := r.devices.Indexes().CreateOne(ctx, mongo.IndexModel{
Keys: bson.D{{Key: "userId", Value: 1}},
}); err != nil {
return err
}
_, err := r.sent.Indexes().CreateOne(ctx, mongo.IndexModel{
Keys:    bson.D{{Key: "createdAt", Value: 1}},
Options: options.Index().SetExpireAfterSeconds(sentLogTTL),
})
return err
}

func (r *NotificationRepository) Upsert(ctx context.Context, d *notification.Device) error {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
now := time.Now().UTC()
_, err := r.devices.UpdateOne(ctx,
bson.M{"_id": d.Token},
bson.M{
"$set":         bson.M{"userId": d.UserID, "platform": d.Platform, "timezone": d.Timezone, "utcOffsetMinutes": d.UTCOffsetMinutes, "updatedAt": now},
"$setOnInsert": bson.M{"createdAt": now},
},
options.UpdateOne().SetUpsert(true),
)
return err
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID string) ([]*notification.Device, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
cur, err := r.devices.Find(ctx, bson.M{"userId": userID})
if err != nil {
return nil, err
}
defer cur.Close(ctx)

out := make([]*notification.Device, 0)
if err := cur.All(ctx, &out); err != nil {
return nil, err
}
return out, nil
}

func (r *NotificationRepository) DeleteToken(ctx context.Context, userID, token string) error {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
_, err := r.devices.DeleteOne(ctx, bson.M{"_id": token, "userId": userID})
return err
}

func (r *NotificationRepository) MarkSent(ctx context.Context, key string) (bool, error) {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
_, err := r.sent.InsertOne(ctx, bson.M{"_id": key, "createdAt": time.Now().UTC()})
if mongo.IsDuplicateKeyError(err) {
return false, nil
}
if err != nil {
return false, err
}
return true, nil
}

func (r *NotificationRepository) ReleaseSent(ctx context.Context, key string) error {
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
_, err := r.sent.DeleteOne(ctx, bson.M{"_id": key})
return err
}