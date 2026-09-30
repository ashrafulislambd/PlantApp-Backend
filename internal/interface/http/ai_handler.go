package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/interface/http/authmw"
	"plantpal-backend/internal/interface/http/reqlocale"
	"plantpal-backend/internal/interface/http/respond"
	chatuc "plantpal-backend/internal/usecase/chat"
	diagnosisuc "plantpal-backend/internal/usecase/diagnosis"
)

// aiHandler serves the Flutter AI Doctor screen's own contract: unversioned,
// authenticated /ai/chat and /ai/diagnose routes returning a flat {text,
// plantName, suggestedChips, diagnosis} shape, unlike every /api/v1/*
// route's {data, error} envelope. It's a thin adapter over the same v1
// usecases, not a separate implementation.
type aiHandler struct {
	chatSvc      *chatuc.Service
	diagnosisSvc *diagnosisuc.Service
}

type aiDiagnosisData struct {
	Issue      string   `json:"issue"`
	Confidence string   `json:"confidence"`
	Severity   string   `json:"severity"`
	Treatment  string   `json:"treatment"`
	Fertilizer string   `json:"fertilizer"`
	ShopItems  []string `json:"shopItems"`
}

type aiReplyResponse struct {
	Text           string           `json:"text"`
	PlantName      *string          `json:"plantName,omitempty"`
	SuggestedChips []string         `json:"suggestedChips"`
	Diagnosis      *aiDiagnosisData `json:"diagnosis,omitempty"`
}

func writeAIReply(w http.ResponseWriter, resp aiReplyResponse) {
	if resp.SuggestedChips == nil {
		resp.SuggestedChips = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

type aiChatRequest struct {
	Text string `json:"text"`
}

// Chat handles POST /ai/chat. The client doesn't maintain a session id of
// its own, so each request gets a fresh ephemeral one — no multi-turn
// context is persisted across calls yet.
func (h *aiHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req aiChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}

	userID, _ := authmw.UserID(r.Context())
	sessionID := fmt.Sprintf("ai-chat-%d", time.Now().UnixNano())
	msgs, err := h.chatSvc.Send(r.Context(), chatuc.SendInput{
		UserID:    userID,
		SessionID: sessionID,
		Content:   req.Text,
		Lang:      reqlocale.Resolve(r),
	})
	if err != nil {
		respond.Error(w, err)
		return
	}

	reply := ""
	if len(msgs) > 0 {
		reply = msgs[len(msgs)-1].Content
	}
	writeAIReply(w, aiReplyResponse{Text: reply})
}

type aiDiagnoseRequest struct {
	Image   []int  `json:"image"`
	Caption string `json:"caption,omitempty"`
}

// Diagnose handles POST /ai/diagnose. The client sends the photo as a raw
// JSON array of byte values (Dio JSON-encoding a Uint8List), not base64 or
// multipart — inefficient for large photos, but this matches what's
// actually sent today.
func (h *aiHandler) Diagnose(w http.ResponseWriter, r *http.Request) {
	var req aiDiagnoseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}
	if len(req.Image) == 0 {
		respond.Error(w, fmt.Errorf("%w: image is required", apperr.ErrInvalidInput))
		return
	}

	imageData := make([]byte, len(req.Image))
	for i, v := range req.Image {
		imageData[i] = byte(v)
	}

	userID, _ := authmw.UserID(r.Context())
	d, err := h.diagnosisSvc.Analyze(r.Context(), diagnosisuc.AnalyzeInput{
		UserID:    userID,
		ImageData: imageData,
	})
	if err != nil {
		respond.Error(w, err)
		return
	}

	localized := d.Localized(reqlocale.Resolve(r))
	writeAIReply(w, aiReplyResponse{
		Text: localized.Issue,
		Diagnosis: &aiDiagnosisData{
			Issue:      localized.Issue,
			Confidence: localized.Confidence,
			Severity:   localized.Severity,
			Treatment:  localized.Cure,
			Fertilizer: localized.Fertilizer,
			ShopItems:  []string{},
		},
	})
}
