package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/chat"
)

func TestChatRepository_ListBySession_OrderedAndScopedToSession(t *testing.T) {
	repo := NewChatRepository()
	ctx := context.Background()
	now := time.Now().UTC()

	second := &chat.Message{ID: "msg_2", UserID: "u1", SessionID: "s1", Role: chat.RoleAssistant, Content: "second", CreatedAt: now}
	first := &chat.Message{ID: "msg_1", UserID: "u1", SessionID: "s1", Role: chat.RoleUser, Content: "first", CreatedAt: now.Add(-time.Minute)}
	otherSession := &chat.Message{ID: "msg_3", UserID: "u1", SessionID: "s2", Role: chat.RoleUser, Content: "other", CreatedAt: now}

	for _, m := range []*chat.Message{second, first, otherSession} {
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	got, err := repo.ListBySession(ctx, "u1", "s1")
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
	got, err := repo.ListBySession(context.Background(), "u1", "unknown")
	if err != nil {
		t.Fatalf("ListBySession() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListBySession(unknown) = %+v, want empty", got)
	}
}

func TestChatRepository_ListSessions_GroupsOrdersAndScopesToUser(t *testing.T) {
	repo := NewChatRepository()
	ctx := context.Background()
	now := time.Now().UTC()

	add := func(id, user, session string, role chat.Role, content string, at time.Time) {
		t.Helper()
		if err := repo.Create(ctx, &chat.Message{ID: id, UserID: user, SessionID: session, Role: role, Content: content, CreatedAt: at}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}
	// s_old: last activity 2h ago. s_new: last activity now. Other user's s_new must not leak.
	add("m1", "u1", "s_old", chat.RoleUser, "  Why are my   leaves yellow?  ", now.Add(-3*time.Hour))
	add("m2", "u1", "s_old", chat.RoleAssistant, "Overwatering, probably.", now.Add(-2*time.Hour))
	add("m3", "u1", "s_new", chat.RoleUser, "Treat leaf spot", now.Add(-time.Minute))
	add("m4", "u1", "s_new", chat.RoleAssistant, "Remove affected leaves.", now)
	add("m5", "u1", "s_new", chat.RoleUser, "thanks", now.Add(-30*time.Second))
	add("m6", "u2", "s_new", chat.RoleUser, "someone else", now)

	got, err := repo.ListSessions(ctx, "u1")
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListSessions() returned %d sessions, want 2", len(got))
	}
	if got[0].ID != "s_new" || got[1].ID != "s_old" {
		t.Errorf("order = [%s, %s], want most recent first [s_new, s_old]", got[0].ID, got[1].ID)
	}
	if got[0].Title != "Treat leaf spot" {
		t.Errorf("title = %q, want the first message", got[0].Title)
	}
	if got[1].Title != "Why are my leaves yellow?" {
		t.Errorf("title = %q, want whitespace collapsed", got[1].Title)
	}
	if got[0].MessageCount != 3 || got[1].MessageCount != 2 {
		t.Errorf("counts = [%d, %d], want [3, 2]", got[0].MessageCount, got[1].MessageCount)
	}
	if !got[0].LastMessageAt.Equal(now) {
		t.Errorf("LastMessageAt = %v, want %v", got[0].LastMessageAt, now)
	}
}

func TestChatRepository_ListSessions_EmptyIsNonNil(t *testing.T) {
	got, err := NewChatRepository().ListSessions(context.Background(), "nobody")
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("ListSessions() = %#v, want empty non-nil slice", got)
	}
}

func TestChatRepository_DeleteSession(t *testing.T) {
	repo := NewChatRepository()
	ctx := context.Background()
	now := time.Now().UTC()
	for _, m := range []*chat.Message{
		{ID: "a", UserID: "u1", SessionID: "s1", Role: chat.RoleUser, Content: "x", CreatedAt: now},
		{ID: "b", UserID: "u1", SessionID: "s2", Role: chat.RoleUser, Content: "y", CreatedAt: now},
		{ID: "c", UserID: "u2", SessionID: "s1", Role: chat.RoleUser, Content: "z", CreatedAt: now},
	} {
		_ = repo.Create(ctx, m)
	}

	if err := repo.DeleteSession(ctx, "u1", "s1"); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if got, _ := repo.ListBySession(ctx, "u1", "s1"); len(got) != 0 {
		t.Errorf("s1 still has %d messages after delete", len(got))
	}
	if got, _ := repo.ListBySession(ctx, "u1", "s2"); len(got) != 1 {
		t.Errorf("deleting s1 must not touch s2")
	}
	if got, _ := repo.ListBySession(ctx, "u2", "s1"); len(got) != 1 {
		t.Errorf("deleting u1's s1 must not touch u2's s1")
	}
	if err := repo.DeleteSession(ctx, "u1", "s1"); !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("second DeleteSession() error = %v, want ErrNotFound", err)
	}
}
