package v1

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/diagnosis"
	"myplantpal-backend/internal/interface/http/authmw"
	"myplantpal-backend/internal/interface/http/reqlocale"
	"myplantpal-backend/internal/interface/http/respond"
	diagnosisuc "myplantpal-backend/internal/usecase/diagnosis"
)

type DiagnosisHandler struct {
	svc *diagnosisuc.Service
}

func NewDiagnosisHandler(svc *diagnosisuc.Service) *DiagnosisHandler {
	return &DiagnosisHandler{svc: svc}
}

type createDiagnosisRequest struct {
	PlantID     *string `json:"plantId,omitempty"`
	ImageBase64 string  `json:"imageBase64"`
}

func (h *DiagnosisHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createDiagnosisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}

	imageData, err := base64.StdEncoding.DecodeString(req.ImageBase64)
	if err != nil {
		respond.Error(w, fmt.Errorf("%w: imageBase64 is not valid base64", apperr.ErrInvalidInput))
		return
	}

	userID, _ := authmw.UserID(r.Context())
	d, err := h.svc.Analyze(r.Context(), diagnosisuc.AnalyzeInput{
		UserID:    userID,
		PlantID:   req.PlantID,
		ImageData: imageData,
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, d.Localized(reqlocale.Resolve(r)))
}

func (h *DiagnosisHandler) List(w http.ResponseWriter, r *http.Request) {
	var plantID *string
	if q := r.URL.Query().Get("plantId"); q != "" {
		plantID = &q
	}
	userID, _ := authmw.UserID(r.Context())
	items, err := h.svc.List(r.Context(), userID, plantID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	lang := reqlocale.Resolve(r)
	localized := make([]diagnosis.Diagnosis, len(items))
	for i, item := range items {
		localized[i] = item.Localized(lang)
	}
	respond.JSON(w, http.StatusOK, localized)
}

func (h *DiagnosisHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	d, err := h.svc.Get(r.Context(), r.PathValue("id"), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, d.Localized(reqlocale.Resolve(r)))
}
