package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/chat"
	"plantpal-backend/internal/interface/http/authmw"
	"plantpal-backend/internal/interface/http/reqlocale"
	"plantpal-backend/internal/interface/http/respond"
	chatuc "plantpal-backend/internal/usecase/chat"
)

type ChatHandler struct {
	svc *chatuc.Service
}

func NewChatHandler(svc *chatuc.Service) *ChatHandler {
	return &ChatHandler{svc: svc}
}

type sendChatRequest struct {
	SessionID string `json:"sessionId"`
	Content   string `json:"content"`
	// DiagnosisID optionally attaches one of the user's scans as context.
	DiagnosisID string `json:"diagnosisId,omitempty"`
}

// Send handles the AI Chat Box's message composer.
func (h *ChatHandler) Send(w http.ResponseWriter, r *http.Request) {
	var req sendChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}

	userID, _ := authmw.UserID(r.Context())
	msgs, err := h.svc.Send(r.Context(), chatuc.SendInput{
		UserID:    userID,
		SessionID: req.SessionID,
		Content:   req.Content,
		Lang:      reqlocale.Resolve(r),

		DiagnosisID: req.DiagnosisID,
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, msgs)
}

// List returns the message history for ?sessionId=.
func (h *ChatHandler) List(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		respond.Error(w, fmt.Errorf("%w: sessionId query param is required", apperr.ErrInvalidInput))
		return
	}
	userID, _ := authmw.UserID(r.Context())
	items, err := h.svc.List(r.Context(), userID, sessionID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, items)
}

// Sessions returns the signed-in user's conversations, most recent first,
// for the chat history menu.
func (h *ChatHandler) Sessions(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	items, err := h.svc.ListSessions(r.Context(), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, items)
}

// DeleteSession deletes one conversation and all of its messages.
func (h *ChatHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	if err := h.svc.DeleteSession(r.Context(), userID, r.PathValue("id")); err != nil {
		respond.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// sseWriteTimeout is how long a single SSE write may take. It replaces the
// server-wide WriteTimeout, which would otherwise cut off any reply that
// takes longer than that in total.
const sseWriteTimeout = 30 * time.Second

// sseStream writes Server-Sent Events. Response headers are only sent with
// the first event (see begin), so a failure before any text exists can
// still be reported as a normal JSON error with a real HTTP status.
type sseStream struct {
	w       http.ResponseWriter
	rc      *http.ResponseController
	started bool
}

func newSSEStream(w http.ResponseWriter) *sseStream {
	return &sseStream{w: w, rc: http.NewResponseController(w)}
}

func (s *sseStream) begin() {
	if s.started {
		return
	}
	s.started = true
	h := s.w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache, no-transform")
	h.Set("X-Accel-Buffering", "no") // tell nginx-style proxies not to buffer
	s.w.WriteHeader(http.StatusOK)
}

// event writes one named event and flushes it to the client immediately.
func (s *sseStream) event(name string, payload any) error {
	s.begin()
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	// Not every ResponseWriter supports deadlines/flushing (tests); that is
	// fine, only real failures matter.
	if err := s.rc.SetWriteDeadline(time.Now().Add(sseWriteTimeout)); err != nil && !errors.Is(err, http.ErrNotSupported) {
		return err
	}
	if _, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", name, b); err != nil {
		return err
	}
	if err := s.rc.Flush(); err != nil && !errors.Is(err, http.ErrNotSupported) {
		return err
	}
	return nil
}

// Stream handles POST /chat/messages/stream: the same contract as Send, but
// the assistant's reply is delivered as Server-Sent Events as it is
// generated:
//
//	event: start  data: {"userMessage": Message, "assistantId": "msg_..."}
//	event: delta  data: {"text": "..."}              (repeated)
//	event: done   data: {"message": Message}         (the stored reply)
//	event: error  data: {"code", "error", "message"?} (partial reply, if any)
//
// Errors that happen before the first delta are plain JSON error responses
// with the usual status codes (400, 429, ...). If the client disconnects,
// the reply generated so far is still stored.
func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request) {
	var req sendChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}

	userID, _ := authmw.UserID(r.Context())
	stream := newSSEStream(w)

	var userMsg *chat.Message
	var assistantID string
	started := false

	assistant, err := h.svc.SendStream(r.Context(), chatuc.SendInput{
		UserID:    userID,
		SessionID: req.SessionID,
		Content:   req.Content,
		Lang:      reqlocale.Resolve(r),

		DiagnosisID: req.DiagnosisID,
	}, chatuc.StreamCallbacks{
		OnStart: func(u *chat.Message, id string) error {
			userMsg, assistantID = u, id
			return nil
		},
		OnDelta: func(text string) error {
			if !started {
				started = true
				if err := stream.event("start", map[string]any{
					"userMessage": userMsg,
					"assistantId": assistantID,
				}); err != nil {
					return err
				}
			}
			return stream.event("delta", map[string]string{"text": text})
		},
	})

	if err != nil {
		if !stream.started {
			respond.Error(w, err)
			return
		}
		log.Printf("chat stream: %v", err)
		code, msg := streamErrorInfo(err)
		payload := map[string]any{"code": code, "error": msg}
		if assistant != nil {
			payload["message"] = assistant
		}
		_ = stream.event("error", payload)
		return
	}

	if !started {
		// Nothing was delivered but no error either (cannot normally
		// happen); still send a well-formed stream.
		_ = stream.event("start", map[string]any{"userMessage": userMsg, "assistantId": assistantID})
	}
	_ = stream.event("done", map[string]any{"message": assistant})
}

// streamErrorInfo maps a mid-stream failure to a stable machine-readable
// code and a message that is safe to show (provider internals stay in the
// server log).
func streamErrorInfo(err error) (code, message string) {
	switch {
	case errors.Is(err, apperr.ErrRateLimited):
		return "rate_limited", "rate limit reached, please wait a moment and try again"
	case errors.Is(err, context.Canceled):
		return "canceled", "the reply was cancelled"
	default:
		return "interrupted", "the reply was interrupted"
	}
}
