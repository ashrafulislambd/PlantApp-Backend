package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/interface/http/respond"
	chatuc "myplantpal-backend/internal/usecase/chat"
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
}

// Send handles the AI Chat Box's message composer.
func (h *ChatHandler) Send(w http.ResponseWriter, r *http.Request) {
	var req sendChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}

	msgs, err := h.svc.Send(r.Context(), chatuc.SendInput{
		SessionID: req.SessionID,
		Content:   req.Content,
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
	items, err := h.svc.List(r.Context(), sessionID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, items)
}
