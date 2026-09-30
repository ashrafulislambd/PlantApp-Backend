// Package notification holds device registrations and the push-sending
// contract used by plant care reminders.
package notification

import (
"context"
"errors"
"time"
)

// ErrInvalidToken is returned by a Sender when the push service says the
// device token is dead (uninstalled app, expired token). The caller should
// forget the token.
var ErrInvalidToken = errors.New("device token is no longer valid")

// Device is one push-capable install of the app. The FCM token is unique per
// install, so it doubles as the primary key.
type Device struct {
Token     string    `json:"token" bson:"_id"`
UserID    string    `json:"-" bson:"userId"`
Platform  string    `json:"platform" bson:"platform"`
CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

type Message struct {
Title string
Body  string
Data  map[string]string
}

// Sender delivers one message to one device token (FCM in production).
type Sender interface {
Send(ctx context.Context, token string, msg Message) error
}

// Repository stores device tokens and a small log of reminders already
// pushed, so a due reminder is delivered once instead of on every tick.
type Repository interface {
// Upsert registers d.Token for d.UserID. Re-registering an existing
// token moves it to the new user (shared phone, new login).
Upsert(ctx context.Context, d *Device) error
ListByUser(ctx context.Context, userID string) ([]*Device, error)
DeleteToken(ctx context.Context, userID, token string) error
// MarkSent records key and reports whether it was new (false = already sent).
MarkSent(ctx context.Context, key string) (bool, error)
// ReleaseSent forgets key so a failed delivery is retried on the next tick.
ReleaseSent(ctx context.Context, key string) error
}