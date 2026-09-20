package memory

import (
	"context"
	"sort"
	"sync"

	"myplantpal-backend/internal/domain/chat"
)

type ChatRepository struct {
	mu    sync.RWMutex
	items map[string][]*chat.Message // keyed by sessionID
}

func NewChatRepository() *ChatRepository {
	return &ChatRepository{items: make(map[string][]*chat.Message)}
}

func (r *ChatRepository) Create(_ context.Context, m *chat.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[m.SessionID] = append(r.items[m.SessionID], m)
	return nil
}

func (r *ChatRepository) ListBySession(_ context.Context, sessionID string) ([]*chat.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msgs := r.items[sessionID]
	out := make([]*chat.Message, len(msgs))
	copy(out, msgs)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
