package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/interface/http/authmw"
	"myplantpal-backend/internal/interface/http/respond"
	plantuc "myplantpal-backend/internal/usecase/plant"
)

type PlantHandler struct {
	svc *plantuc.Service
}

func NewPlantHandler(svc *plantuc.Service) *PlantHandler {
	return &PlantHandler{svc: svc}
}

type createPlantRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	AgeStage string `json:"ageStage"`
}

// Create handles "Create My Roadmap" on the Maintainance screen.
func (h *PlantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createPlantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}

	userID, _ := authmw.UserID(r.Context())
	p, err := h.svc.Create(r.Context(), plantuc.CreateInput{
		UserID:   userID,
		Name:     req.Name,
		Type:     req.Type,
		AgeStage: req.AgeStage,
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, p)
}

func (h *PlantHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, items)
}

func (h *PlantHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	p, err := h.svc.Get(r.Context(), r.PathValue("id"), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

func (h *PlantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	if err := h.svc.Delete(r.Context(), r.PathValue("id"), userID); err != nil {
		respond.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
