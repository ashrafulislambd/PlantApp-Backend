package chat

import "context"

// Repository persists chat Messages. The MongoDB implementation lives in
// internal/infrastructure/repository/mongo; an in-memory implementation
// also exists in internal/infrastructure/repository/memory for local dev
// without a database.
//
// ListBySession is scoped to a single owning userID: a session ID that
// happens to collide with another user's only ever returns that user's own
// messages.
type Repository interface {
	Create(ctx context.Context, m *Message) error
	ListBySession(ctx context.Context, userID, sessionID string) ([]*Message, error)
}
