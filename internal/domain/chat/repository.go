package chat

import "context"

// Repository persists chat Messages.
type Repository interface {
	Create(ctx context.Context, m *Message) error
	ListBySession(ctx context.Context, sessionID string) ([]*Message, error)
}
