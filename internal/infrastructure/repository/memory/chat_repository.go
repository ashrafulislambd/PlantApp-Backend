package memory

import (
	"context"
	"sort"
	"strings"
	"sync"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/chat"
)

type ChatRepository struct {
	mu    sync.RWMutex
	items map[string][]*chat.Message // keyed by userID+"|"+sessionID
}

func NewChatRepository() *ChatRepository {
	return &ChatRepository{items: make(map[string][]*chat.Message)}
}

func chatKey(userID, sessionID string) string {
	return userID + "|" + sessionID
}

func (r *ChatRepository) Create(_ context.Context, m *chat.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := chatKey(m.UserID, m.SessionID)
	r.items[key] = append(r.items[key], m)
	return nil
}

func (r *ChatRepository) ListBySession(_ context.Context, userID, sessionID string) ([]*chat.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msgs := r.items[chatKey(userID, sessionID)]
	out := make([]*chat.Message, len(msgs))
	copy(out, msgs)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *ChatRepository) ListSessions(_ context.Context, userID string) ([]*chat.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefix := userID + "|"
	out := make([]*chat.Session, 0)
	for key, msgs := range r.items {
		if !strings.HasPrefix(key, prefix) || len(msgs) == 0 {
			continue
		}
		first, last := msgs[0], msgs[0]
		for _, m := range msgs {
			if m.CreatedAt.Before(first.CreatedAt) {
				first = m
			}
			if m.CreatedAt.After(last.CreatedAt) {
				last = m
			}
		}
		out = append(out, &chat.Session{
			ID:            first.SessionID,
			Title:         chat.SessionTitle(first.Content, first.DiagnosisID),
			LastMessageAt: last.CreatedAt,
			MessageCount:  len(msgs),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastMessageAt.After(out[j].LastMessageAt) })
	return out, nil
}

func (r *ChatRepository) DeleteSession(_ context.Context, userID, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := chatKey(userID, sessionID)
	if len(r.items[key]) == 0 {
		return apperr.ErrNotFound
	}
	delete(r.items, key)
	return nil
}
