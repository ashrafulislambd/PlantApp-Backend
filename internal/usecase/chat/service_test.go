package chat

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"plantpal-backend/internal/domain/aiprovider"
	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/chat"
	"plantpal-backend/internal/domain/diagnosis"
	"plantpal-backend/internal/idgen"
	"plantpal-backend/internal/infrastructure/repository/memory"
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

// streamStub is a chat.ReplyProvider that also streams its pieces.
type streamStub struct {
	pieces  []string
	failAt  int // fail after this many pieces were sent (0 = never)
	failErr error
}

func (p streamStub) Reply(context.Context, []*chat.Message, string, string) (chat.ReplyResult, error) {
	return chat.ReplyResult{Text: strings.Join(p.pieces, ""), Provider: aiprovider.Mock}, nil
}

func (p streamStub) ReplyStream(_ context.Context, _ []*chat.Message, _ string, _ string, onDelta func(string) error) (chat.ReplyResult, error) {
	var sent strings.Builder
	for i, piece := range p.pieces {
		if p.failAt > 0 && i == p.failAt {
			return chat.ReplyResult{Text: sent.String(), Provider: aiprovider.Mock}, p.failErr
		}
		sent.WriteString(piece)
		if err := onDelta(piece); err != nil {
			return chat.ReplyResult{Text: sent.String(), Provider: aiprovider.Mock}, err
		}
	}
	return chat.ReplyResult{Text: sent.String(), Provider: aiprovider.Mock}, nil
}

func TestSendStream_StreamsAndStoresBothMessages(t *testing.T) {
	svc := newTestService(streamStub{pieces: []string{"Hel", "lo ", "there"}})
	var got []string
	var startUser *chat.Message
	var startID string

	assistant, err := svc.SendStream(context.Background(),
		SendInput{UserID: "u1", SessionID: "s1", Content: "hi"},
		StreamCallbacks{
			OnStart: func(u *chat.Message, id string) error { startUser, startID = u, id; return nil },
			OnDelta: func(s string) error { got = append(got, s); return nil },
		})
	if err != nil {
		t.Fatalf("SendStream() error = %v", err)
	}
	if strings.Join(got, "|") != "Hel|lo |there" {
		t.Errorf("deltas = %q", got)
	}
	if startUser == nil || startUser.Content != "hi" || startUser.Role != chat.RoleUser {
		t.Errorf("OnStart user message = %+v", startUser)
	}
	if assistant.ID != startID || assistant.Content != "Hello there" {
		t.Errorf("assistant = %+v, want id %q and full text", assistant, startID)
	}
	hist, _ := svc.List(context.Background(), "u1", "s1")
	if len(hist) != 2 || hist[0].Role != chat.RoleUser || hist[1].Content != "Hello there" {
		t.Errorf("history = %+v", hist)
	}
}

func TestSendStream_KeepsPartialReplyWhenProviderFailsMidStream(t *testing.T) {
	boom := errors.New("boom")
	svc := newTestService(streamStub{pieces: []string{"one ", "two ", "three"}, failAt: 2, failErr: boom})

	assistant, err := svc.SendStream(context.Background(),
		SendInput{UserID: "u1", SessionID: "s1", Content: "hi"}, StreamCallbacks{})
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want %v", err, boom)
	}
	if assistant == nil || assistant.Content != "one two " {
		t.Fatalf("assistant = %+v, want the partial reply", assistant)
	}
	hist, _ := svc.List(context.Background(), "u1", "s1")
	if len(hist) != 2 || hist[1].Content != "one two " {
		t.Errorf("history = %+v, want user + partial assistant", hist)
	}
}

func TestSendStream_StoresPartialEvenWhenRequestContextCancelled(t *testing.T) {
	svc := newTestService(streamStub{pieces: []string{"a", "b", "c"}})
	ctx, cancel := context.WithCancel(context.Background())

	// The client "presses stop" after the first piece.
	assistant, err := svc.SendStream(ctx,
		SendInput{UserID: "u1", SessionID: "s1", Content: "hi"},
		StreamCallbacks{OnDelta: func(string) error { cancel(); return ctx.Err() }})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if assistant == nil || assistant.Content != "a" {
		t.Fatalf("assistant = %+v, want partial %q", assistant, "a")
	}
	hist, _ := svc.List(context.Background(), "u1", "s1")
	if len(hist) != 2 {
		t.Errorf("history has %d messages, want 2 (user + partial)", len(hist))
	}
}

func TestSendStream_NoAssistantMessageWhenNothingWasProduced(t *testing.T) {
	boom := errors.New("down")
	svc := newTestService(stubReplyProvider{err: boom}) // cannot stream; Reply fails
	assistant, err := svc.SendStream(context.Background(),
		SendInput{UserID: "u1", SessionID: "s1", Content: "hi"}, StreamCallbacks{})
	if !errors.Is(err, boom) || assistant != nil {
		t.Fatalf("assistant = %+v, err = %v; want nil, %v", assistant, err, boom)
	}
	hist, _ := svc.List(context.Background(), "u1", "s1")
	if len(hist) != 1 {
		t.Errorf("history has %d messages, want only the user's", len(hist))
	}
}

func TestSendStream_FallsBackToSinglePieceForNonStreamingProvider(t *testing.T) {
	svc := newTestService(stubReplyProvider{reply: "whole answer"})
	var got []string
	assistant, err := svc.SendStream(context.Background(),
		SendInput{UserID: "u1", SessionID: "s1", Content: "hi"},
		StreamCallbacks{OnDelta: func(s string) error { got = append(got, s); return nil }})
	if err != nil || assistant.Content != "whole answer" || len(got) != 1 {
		t.Fatalf("assistant = %+v, err = %v, deltas = %q", assistant, err, got)
	}
}

func TestSendStream_ValidatesInput(t *testing.T) {
	svc := newTestService(streamStub{pieces: []string{"x"}})
	for _, in := range []SendInput{
		{UserID: "u1", SessionID: "", Content: "hi"},
		{UserID: "u1", SessionID: "s1", Content: " "},
		{UserID: "", SessionID: "s1", Content: "hi"},
	} {
		if _, err := svc.SendStream(context.Background(), in, StreamCallbacks{}); !errors.Is(err, apperr.ErrInvalidInput) {
			t.Errorf("SendStream(%+v) error = %v, want ErrInvalidInput", in, err)
		}
	}
}

func TestListAndDeleteSessions(t *testing.T) {
	svc := newTestService(stubReplyProvider{reply: "ok"})
	ctx := context.Background()
	_, _ = svc.Send(ctx, SendInput{UserID: "u1", SessionID: "s1", Content: "first question"})
	_, _ = svc.Send(ctx, SendInput{UserID: "u1", SessionID: "s2", Content: "second question"})

	sessions, err := svc.ListSessions(ctx, "u1")
	if err != nil || len(sessions) != 2 {
		t.Fatalf("ListSessions() = %+v, %v; want 2 sessions", sessions, err)
	}
	if err := svc.DeleteSession(ctx, "u1", "s1"); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if sessions, _ = svc.ListSessions(ctx, "u1"); len(sessions) != 1 || sessions[0].ID != "s2" {
		t.Errorf("after delete sessions = %+v, want only s2", sessions)
	}
	if err := svc.DeleteSession(ctx, "u1", "s1"); !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("deleting again error = %v, want ErrNotFound", err)
	}
	if err := svc.DeleteSession(ctx, "u2", "s2"); !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("deleting another user's session error = %v, want ErrNotFound", err)
	}
	if _, err := svc.ListSessions(ctx, ""); !errors.Is(err, apperr.ErrInvalidInput) {
		t.Errorf("ListSessions(\"\") error = %v, want ErrInvalidInput", err)
	}
}

// --- scans in chat -----------------------------------------------------

// capturingProvider records what the model would have been sent.
type capturingProvider struct {
	history []*chat.Message
	prompt  string
}

func (p *capturingProvider) Reply(_ context.Context, h []*chat.Message, prompt string, _ string) (chat.ReplyResult, error) {
	p.history, p.prompt = h, prompt
	return chat.ReplyResult{Text: "ok", Provider: aiprovider.Mock}, nil
}

func newScanService(t *testing.T) (*Service, *capturingProvider, *diagnosis.Diagnosis) {
	t.Helper()
	diagRepo := memory.NewDiagnosisRepository()
	d := &diagnosis.Diagnosis{
		ID: "diag_1", UserID: "u1", Issue: "Leaf spot", Cure: "Remove leaves",
		IssueBn: "পাতায় দাগ", CureBn: "পাতা সরান", Severity: "Mild", Provider: "mock",
	}
	if err := diagRepo.Create(context.Background(), d); err != nil {
		t.Fatalf("seed diagnosis: %v", err)
	}
	p := &capturingProvider{}
	svc := NewService(memory.NewChatRepository(), p, idgen.New())
	svc.SetScanLookup(diagRepo)
	return svc, p, d
}

func TestSend_RejectsAnotherUsersOrUnknownDiagnosisID(t *testing.T) {
	svc, _, d := newScanService(t)
	for _, in := range []SendInput{
		{UserID: "u2", SessionID: "s1", Content: "hi", DiagnosisID: d.ID}, // someone else's scan
		{UserID: "u1", SessionID: "s1", Content: "hi", DiagnosisID: "nope"},
	} {
		if _, err := svc.Send(context.Background(), in); !errors.Is(err, apperr.ErrInvalidInput) {
			t.Errorf("Send(%+v) error = %v, want ErrInvalidInput", in, err)
		}
	}
	if list, _ := svc.List(context.Background(), "u2", "s1"); len(list) != 0 {
		t.Errorf("rejected message was stored: %+v", list)
	}
}

func TestSend_StoresDiagnosisIDAndPromptsWithScanResultOnce(t *testing.T) {
	svc, p, d := newScanService(t)
	ctx := context.Background()

	msgs, err := svc.Send(ctx, SendInput{UserID: "u1", SessionID: "s1", Content: "What should I do?", DiagnosisID: d.ID})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if msgs[0].DiagnosisID != d.ID || msgs[0].Content != "What should I do?" {
		t.Errorf("user turn = %+v, want diagnosisId set and the raw text stored", msgs[0])
	}
	if !strings.Contains(p.prompt, "[Scan result: issue Leaf spot; cure Remove leaves") || !strings.HasSuffix(p.prompt, "What should I do?") {
		t.Errorf("prompt = %q, want scan result then the question", p.prompt)
	}

	// A later turn without diagnosisId: the session still remembers the scan.
	if _, err := svc.Send(ctx, SendInput{UserID: "u1", SessionID: "s1", Content: "And watering?"}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if p.prompt != "And watering?" {
		t.Errorf("prompt = %q, want the plain text (scan already described in history)", p.prompt)
	}
	if len(p.history) != 2 || !strings.Contains(p.history[0].Content, "[Scan result: issue Leaf spot") {
		t.Errorf("history = %+v, want first turn rewritten with the scan result", p.history)
	}
	// Stored history must be untouched by prompt rewriting.
	stored, _ := svc.List(ctx, "u1", "s1")
	if stored[0].Content != "What should I do?" {
		t.Errorf("stored content = %q, rewriting leaked into storage", stored[0].Content)
	}
}

func TestRecordScan_SavesOrderedTurnsAndMessageJSONHasImageURL(t *testing.T) {
	svc, p, d := newScanService(t)
	ctx := context.Background()

	if err := svc.RecordScan(ctx, "u1", "s1", "  is this mold?  ", d, "bn"); err != nil {
		t.Fatalf("RecordScan() error = %v", err)
	}
	got, _ := svc.List(ctx, "u1", "s1")
	if len(got) != 2 || got[0].Role != chat.RoleUser || got[1].Role != chat.RoleAssistant {
		t.Fatalf("turns = %+v, want user then assistant", got)
	}
	if got[0].Content != "is this mold?" || got[0].DiagnosisID != d.ID || got[1].DiagnosisID != d.ID {
		t.Errorf("turns = %+v, want caption + diagnosisId on both", got)
	}
	if !strings.Contains(got[1].Content, "পাতায় দাগ") {
		t.Errorf("assistant summary = %q, want the Bengali issue", got[1].Content)
	}
	b, _ := json.Marshal(got[0])
	if !strings.Contains(string(b), `"diagnosisId":"diag_1"`) || !strings.Contains(string(b), `"imageUrl":"/api/v1/diagnoses/diag_1/image"`) {
		t.Errorf("json = %s, want diagnosisId and imageUrl", b)
	}

	// Follow-up: scan described once in the model history (user turn), the
	// assistant turn keeps its own summary.
	if _, err := svc.Send(ctx, SendInput{UserID: "u1", SessionID: "s1", Content: "ok thanks"}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(p.history[0].Content, "[Scan result:") || strings.Contains(p.history[1].Content, "[Scan result:") {
		t.Errorf("history = %q / %q", p.history[0].Content, p.history[1].Content)
	}
}

func TestRecordScan_CaptionlessPhotoGetsPhotoTitleAndNonEmptyPrompt(t *testing.T) {
	svc, p, d := newScanService(t)
	ctx := context.Background()
	if err := svc.RecordScan(ctx, "u1", "s1", "", d, "en"); err != nil {
		t.Fatal(err)
	}
	sessions, _ := svc.ListSessions(ctx, "u1")
	if len(sessions) != 1 || sessions[0].Title != chat.PhotoTitle {
		t.Errorf("sessions = %+v, want the photo title", sessions)
	}
	if _, err := svc.Send(ctx, SendInput{UserID: "u1", SessionID: "s1", Content: "hi"}); err != nil {
		t.Fatal(err)
	}
	for _, m := range p.history {
		if strings.TrimSpace(m.Content) == "" {
			t.Errorf("empty message reached the model: %+v", m)
		}
	}
}
