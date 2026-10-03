// Package chat implements the application logic behind the AI Chat Box:
// post a message, get an assistant reply, keep per-session history.
package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/chat"
	"plantpal-backend/internal/domain/diagnosis"
	"plantpal-backend/internal/idgen"
)

// ScanLookup finds one of a user's saved scans. It is satisfied by
// diagnosis.Repository, whose GetByID is scoped to the owning user: another
// user's scan is reported as apperr.ErrNotFound.
type ScanLookup interface {
	GetByID(ctx context.Context, id, userID string) (*diagnosis.Diagnosis, error)
}

type Service struct {
	repo     chat.Repository
	provider chat.ReplyProvider
	ids      idgen.Generator
	scans    ScanLookup // optional; nil disables diagnosisId
}

func NewService(repo chat.Repository, provider chat.ReplyProvider, ids idgen.Generator) *Service {
	return &Service{repo: repo, provider: provider, ids: ids}
}

// SetScanLookup enables attaching a saved scan (SendInput.DiagnosisID) to a
// message and remembering scans in the model prompt.
func (s *Service) SetScanLookup(l ScanLookup) { s.scans = l }

type SendInput struct {
	UserID    string
	SessionID string
	Content   string
	Lang      string

	// DiagnosisID optionally attaches one of the user's saved scans as
	// context for this message.
	DiagnosisID string
}

// turn is a validated request, ready to store and to send to the model.
type turn struct {
	sessionID   string
	content     string
	diagnosisID string
	// history is the session so far, with scan turns rewritten as text the
	// model understands (see promptHistory).
	history []*chat.Message
	// prompt is the new message as the model sees it: content, preceded by
	// the scan result when this is the first message to mention that scan.
	prompt string
}

// prepare validates in, checks diagnosisId belongs to the user and builds
// what the model will see. Nothing is stored yet.
func (s *Service) prepare(ctx context.Context, in SendInput) (*turn, error) {
	sessionID := strings.TrimSpace(in.SessionID)
	content := strings.TrimSpace(in.Content)
	if sessionID == "" || content == "" {
		return nil, fmt.Errorf("%w: sessionId and content are required", apperr.ErrInvalidInput)
	}
	if in.UserID == "" {
		return nil, fmt.Errorf("%w: userID is required", apperr.ErrInvalidInput)
	}

	diagnosisID := strings.TrimSpace(in.DiagnosisID)
	var scan *diagnosis.Diagnosis
	if diagnosisID != "" {
		if s.scans == nil {
			return nil, fmt.Errorf("%w: diagnosisId is not supported", apperr.ErrInvalidInput)
		}
		d, err := s.scans.GetByID(ctx, diagnosisID, in.UserID)
		if errors.Is(err, apperr.ErrNotFound) {
			// Same answer for "does not exist" and "belongs to someone
			// else", so ids cannot be probed.
			return nil, fmt.Errorf("%w: unknown diagnosisId", apperr.ErrInvalidInput)
		}
		if err != nil {
			return nil, err
		}
		scan = d
	}

	history, err := s.repo.ListBySession(ctx, in.UserID, sessionID)
	if err != nil {
		return nil, err
	}
	promptHist, seen := s.promptHistory(ctx, in.UserID, history)

	prompt := content
	if scan != nil && !seen[scan.ID] {
		prompt = scan.PromptText() + "\n" + content
	}
	return &turn{
		sessionID:   sessionID,
		content:     content,
		diagnosisID: diagnosisID,
		history:     promptHist,
		prompt:      prompt,
	}, nil
}

// promptHistory returns history ready for the model: every message that
// carries a diagnosisId is rewritten as text, so the model remembers the
// scan on every later turn. The first message that mentions a scan gets the
// "[Scan result: ...]" line; later ones keep their own text (or a short
// "photo attached" marker when they have none), so the result is not
// repeated. It also reports which scans it has already described.
func (s *Service) promptHistory(ctx context.Context, userID string, history []*chat.Message) ([]*chat.Message, map[string]bool) {
	seen := map[string]bool{}
	cache := map[string]*diagnosis.Diagnosis{}
	out := make([]*chat.Message, len(history))

	for i, m := range history {
		if m.DiagnosisID == "" {
			out[i] = m
			continue
		}
		cp := *m
		text := strings.TrimSpace(m.Content)

		if !seen[m.DiagnosisID] {
			d, ok := cache[m.DiagnosisID]
			if !ok && s.scans != nil {
				if found, err := s.scans.GetByID(ctx, m.DiagnosisID, userID); err == nil {
					d = found
				}
				cache[m.DiagnosisID] = d
			}
			if d != nil {
				seen[m.DiagnosisID] = true
				if text == "" {
					text = d.PromptText()
				} else {
					text = d.PromptText() + "\n" + text
				}
			}
		}
		if text == "" {
			text = "[The user attached a photo of their plant]"
		}
		cp.Content = text
		out[i] = &cp
	}
	return out, seen
}

// Send stores the user's message, generates an assistant reply, stores
// that too, and returns both in chronological order.
func (s *Service) Send(ctx context.Context, in SendInput) ([]*chat.Message, error) {
	t, err := s.prepare(ctx, in)
	if err != nil {
		return nil, err
	}

	userMsg := &chat.Message{
		ID:          s.ids.New("msg"),
		UserID:      in.UserID,
		SessionID:   t.sessionID,
		Role:        chat.RoleUser,
		Content:     t.content,
		CreatedAt:   time.Now().UTC(),
		DiagnosisID: t.diagnosisID,
	}
	if err := s.repo.Create(ctx, userMsg); err != nil {
		return nil, err
	}

	result, err := s.provider.Reply(ctx, t.history, t.prompt, in.Lang)
	if err != nil {
		return nil, err
	}

	assistantMsg := &chat.Message{
		ID:        s.ids.New("msg"),
		UserID:    in.UserID,
		SessionID: t.sessionID,
		Role:      chat.RoleAssistant,
		Content:   result.Text,
		CreatedAt: time.Now().UTC(),
		Provider:  string(result.Provider),
	}
	if err := s.repo.Create(ctx, assistantMsg); err != nil {
		return nil, err
	}

	return []*chat.Message{userMsg, assistantMsg}, nil
}

func (s *Service) List(ctx context.Context, userID, sessionID string) ([]*chat.Message, error) {
	return s.repo.ListBySession(ctx, userID, sessionID)
}

// ListSessions returns the user's conversations, most recent first.
func (s *Service) ListSessions(ctx context.Context, userID string) ([]*chat.Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: userID is required", apperr.ErrInvalidInput)
	}
	return s.repo.ListSessions(ctx, userID)
}

// DeleteSession removes one of the user's conversations.
func (s *Service) DeleteSession(ctx context.Context, userID, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if userID == "" || sessionID == "" {
		return fmt.Errorf("%w: userID and sessionID are required", apperr.ErrInvalidInput)
	}
	return s.repo.DeleteSession(ctx, userID, sessionID)
}

// StreamCallbacks receive the live parts of a streamed reply.
type StreamCallbacks struct {
	// OnStart is called once, after the user's message is stored and before
	// the first OnDelta. assistantID is the ID the final assistant message
	// will have, so a client can key its in-progress bubble on it.
	OnStart func(userMsg *chat.Message, assistantID string) error
	// OnDelta is called with each new piece of reply text. Returning an
	// error aborts the stream (the partial reply is still saved).
	OnDelta func(text string) error
}

// SendStream is Send with an incremental reply: it stores the user's
// message, streams the assistant's reply through cb, then stores the full
// reply.
//
// If the reply is cut short after some text was produced (the client
// disconnected, the user pressed stop, or the provider failed mid-way), the
// partial text is still stored so the history matches what the user saw, and
// the assistant message is returned together with the error.
//
// If nothing was produced, no assistant message is stored and a nil message
// is returned with the error.
func (s *Service) SendStream(ctx context.Context, in SendInput, cb StreamCallbacks) (*chat.Message, error) {
	t, err := s.prepare(ctx, in)
	if err != nil {
		return nil, err
	}
	sessionID, content, history := t.sessionID, t.content, t.history

	userMsg := &chat.Message{
		ID:          s.ids.New("msg"),
		UserID:      in.UserID,
		SessionID:   sessionID,
		Role:        chat.RoleUser,
		Content:     content,
		CreatedAt:   time.Now().UTC(),
		DiagnosisID: t.diagnosisID,
	}
	if err := s.repo.Create(ctx, userMsg); err != nil {
		return nil, err
	}

	assistantID := s.ids.New("msg")
	if cb.OnStart != nil {
		if err := cb.OnStart(userMsg, assistantID); err != nil {
			return nil, err
		}
	}
	onDelta := cb.OnDelta
	if onDelta == nil {
		onDelta = func(string) error { return nil }
	}

	var result chat.ReplyResult
	if sp, ok := s.provider.(chat.StreamProvider); ok {
		result, err = sp.ReplyStream(ctx, history, t.prompt, in.Lang, onDelta)
	} else {
		// Provider cannot stream: deliver its whole reply as one piece.
		result, err = s.provider.Reply(ctx, history, t.prompt, in.Lang)
		if err == nil {
			err = onDelta(result.Text)
		}
	}

	text := result.Text
	if err != nil && strings.TrimSpace(text) == "" {
		return nil, err
	}

	assistantMsg := &chat.Message{
		ID:        assistantID,
		UserID:    in.UserID,
		SessionID: sessionID,
		Role:      chat.RoleAssistant,
		Content:   text,
		CreatedAt: time.Now().UTC(),
		Provider:  string(result.Provider),
	}
	// The request context may already be cancelled (that is how a client
	// stop reaches us); the reply must still be saved.
	if cerr := s.repo.Create(context.WithoutCancel(ctx), assistantMsg); cerr != nil {
		return nil, cerr
	}
	return assistantMsg, err
}

// RecordScan saves a photo scan made from the chat as two turns of the
// session: the user's photo with their caption, and a short assistant
// summary of the result. Both reference the already-saved Diagnosis (and so
// its photo) by id; nothing is stored twice. It implements
// diagnosisuc.ChatRecorder.
func (s *Service) RecordScan(ctx context.Context, userID, sessionID, note string, d *diagnosis.Diagnosis, lang string) error {
	sessionID = strings.TrimSpace(sessionID)
	if userID == "" || sessionID == "" || d == nil {
		return fmt.Errorf("%w: userID, sessionID and diagnosis are required", apperr.ErrInvalidInput)
	}

	now := time.Now().UTC()
	userMsg := &chat.Message{
		ID:          s.ids.New("msg"),
		UserID:      userID,
		SessionID:   sessionID,
		Role:        chat.RoleUser,
		Content:     strings.TrimSpace(note),
		CreatedAt:   now,
		DiagnosisID: d.ID,
	}
	// Later than the user turn, so history sorts them in the right order
	// even at millisecond timestamp precision.
	assistantMsg := &chat.Message{
		ID:          s.ids.New("msg"),
		UserID:      userID,
		SessionID:   sessionID,
		Role:        chat.RoleAssistant,
		Content:     d.ChatSummary(lang),
		CreatedAt:   now.Add(2 * time.Millisecond),
		Provider:    d.Provider,
		DiagnosisID: d.ID,
	}
	if err := s.repo.Create(ctx, userMsg); err != nil {
		return err
	}
	return s.repo.Create(ctx, assistantMsg)
}
