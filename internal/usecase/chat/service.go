// Package chat implements the application logic behind the AI Chat Box:
// post a message, get an assistant reply, keep per-session history.
package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/chat"
	"myplantpal-backend/internal/idgen"
)

type Service struct {
	repo     chat.Repository
	provider chat.ReplyProvider
	ids      idgen.Generator
}

func NewService(repo chat.Repository, provider chat.ReplyProvider, ids idgen.Generator) *Service {
	return &Service{repo: repo, provider: provider, ids: ids}
}

type SendInput struct {
	SessionID string
	Content   string
}

// Send stores the user's message, generates an assistant reply, stores
// that too, and returns both in chronological order.
func (s *Service) Send(ctx context.Context, in SendInput) ([]*chat.Message, error) {
	sessionID := strings.TrimSpace(in.SessionID)
	content := strings.TrimSpace(in.Content)
	if sessionID == "" || content == "" {
		return nil, fmt.Errorf("%w: sessionId and content are required", apperr.ErrInvalidInput)
	}

	history, err := s.repo.ListBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	userMsg := &chat.Message{
		ID:        s.ids.New("msg"),
		SessionID: sessionID,
		Role:      chat.RoleUser,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, userMsg); err != nil {
		return nil, err
	}

	replyText, err := s.provider.Reply(ctx, history, content)
	if err != nil {
		return nil, err
	}

	assistantMsg := &chat.Message{
		ID:        s.ids.New("msg"),
		SessionID: sessionID,
		Role:      chat.RoleAssistant,
		Content:   replyText,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, assistantMsg); err != nil {
		return nil, err
	}

	return []*chat.Message{userMsg, assistantMsg}, nil
}

func (s *Service) List(ctx context.Context, sessionID string) ([]*chat.Message, error) {
	return s.repo.ListBySession(ctx, sessionID)
}
