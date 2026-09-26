package chat

import (
	"context"
	"errors"
	"testing"

	"myplantpal-backend/internal/domain/aiprovider"
	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/chat"
	"myplantpal-backend/internal/idgen"
	"myplantpal-backend/internal/infrastructure/repository/memory"
)

type stubReplyProvider struct {
	reply string
	err   error
}

func (p stubReplyProvider) Reply(_ context.Context, _ []*chat.Message, _ string, _ string) (chat.ReplyResult, error) {
	return chat.ReplyResult{Text: p.reply, Provider: aiprovider.Mock}, p.err
}

func newTestService(provider chat.ReplyProvider) *Service {
	return NewService(memory.NewChatRepository(), provider, idgen.New())
}

func TestSend_RequiresSessionIDAndContent(t *testing.T) {
	svc := newTestService(stubReplyProvider{reply: "hi"})
	ctx := context.Background()

	cases := []SendInput{
		{UserID: "user_1", SessionID: "", Content: "hello"},
		{UserID: "user_1", SessionID: "s1", Content: "  "},
		{UserID: "", SessionID: "s1", Content: "hello"},
	}
	for _, in := range cases {
		if _, err := svc.Send(ctx, in); !errors.Is(err, apperr.ErrInvalidInput) {
			t.Errorf("Send(%+v) error = %v, want %v", in, err, apperr.ErrInvalidInput)
		}
	}
}

func TestSend_StoresUserAndAssistantMessagesInOrder(t *testing.T) {
	svc := newTestService(stubReplyProvider{reply: "an assistant reply"})
	ctx := context.Background()

	msgs, err := svc.Send(ctx, SendInput{UserID: "user_1", SessionID: "s1", Content: "hello"})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("Send() returned %d messages, want 2", len(msgs))
	}
	if msgs[0].Role != chat.RoleUser || msgs[0].Content != "hello" {
		t.Errorf("msgs[0] = %+v, want user message with content %q", msgs[0], "hello")
	}
	if msgs[1].Role != chat.RoleAssistant || msgs[1].Content != "an assistant reply" {
		t.Errorf("msgs[1] = %+v, want assistant message with content %q", msgs[1], "an assistant reply")
	}

	history, err := svc.List(ctx, "user_1", "s1")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(history) != 2 {
		t.Errorf("List() returned %d messages, want 2", len(history))
	}
}

func TestSend_PropagatesProviderError(t *testing.T) {
	wantErr := errors.New("provider exploded")
	svc := newTestService(stubReplyProvider{err: wantErr})
	_, err := svc.Send(context.Background(), SendInput{UserID: "user_1", SessionID: "s1", Content: "hello"})
	if !errors.Is(err, wantErr) {
		t.Errorf("Send() error = %v, want %v", err, wantErr)
	}
}

func TestList_UnknownSessionReturnsEmpty(t *testing.T) {
	svc := newTestService(stubReplyProvider{reply: "hi"})
	list, err := svc.List(context.Background(), "user_1", "unknown")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 0 {
		t.Errorf("List(unknown) = %+v, want empty", list)
	}
}
