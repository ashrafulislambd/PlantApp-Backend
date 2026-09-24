package memory

import (
	"context"
	"testing"
	"time"

	"myplantpal-backend/internal/domain/chat"
)

func TestChatRepository_ListBySession_OrderedAndScopedToSession(t *testing.T) {
	repo := NewChatRepository()
	ctx := context.Background()
	now := time.Now().UTC()

	second := &chat.Message{ID: "msg_2", SessionID: "s1", Role: chat.RoleAssistant, Content: "second", CreatedAt: now}
	first := &chat.Message{ID: "msg_1", SessionID: "s1", Role: chat.RoleUser, Content: "first", CreatedAt: now.Add(-time.Minute)}
	otherSession := &chat.Message{ID: "msg_3", SessionID: "s2", Role: chat.RoleUser, Content: "other", CreatedAt: now}

	for _, m := range []*chat.Message{second, first, otherSession} {
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	got, err := repo.ListBySession(ctx, "s1")
	if err != nil {
		t.Fatalf("ListBySession() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListBySession(s1) returned %d messages, want 2", len(got))
	}
	if got[0].ID != "msg_1" || got[1].ID != "msg_2" {
		t.Errorf("ListBySession(s1) not chronologically ordered: got [%s, %s]", got[0].ID, got[1].ID)
	}
}

func TestChatRepository_ListBySession_UnknownSessionReturnsEmpty(t *testing.T) {
	repo := NewChatRepository()
	got, err := repo.ListBySession(context.Background(), "unknown")
	if err != nil {
		t.Fatalf("ListBySession() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListBySession(unknown) = %+v, want empty", got)
	}
}
