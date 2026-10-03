package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/interface/http/authmw"
	"plantpal-backend/internal/interface/http/reqlocale"
	"plantpal-backend/internal/interface/http/respond"
	plantuc "plantpal-backend/internal/usecase/plant"
)

type PlantHandler struct {
	svc *plantuc.Service
}

func NewPlantHandler(svc *plantuc.Service) *PlantHandler {
	return &PlantHandler{svc: svc}
}

// createPlantRequest's JSON keys match the Flutter client's NewPlant/
// PlantDto exactly (see lib/features/plants/data/models/plant_dto.dart's
// newPlantToJson) - nickname/species, not name/type.
type createPlantRequest struct {
	Name                  string     `json:"nickname"`
	Type                  string     `json:"species"`
	AgeStage              string     `json:"ageStage"`
	Location              string     `json:"location"`
	Sunlight              string     `json:"sunlight"`
	Image                 string     `json:"image"`
	WateringFrequencyDays int        `json:"wateringFrequencyDays"`
	LastWatered           *time.Time `json:"lastWatered"`
}

func (h *PlantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createPlantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}

	userID, _ := authmw.UserID(r.Context())
	p, err := h.svc.Create(r.Context(), plantuc.CreateInput{
		UserID:                userID,
		Name:                  req.Name,
		Type:                  req.Type,
		AgeStage:              req.AgeStage,
		Location:              req.Location,
		Sunlight:              req.Sunlight,
		Image:                 req.Image,
		WateringFrequencyDays: req.WateringFrequencyDays,
		LastWatered:           req.LastWatered,
		Lang:                  reqlocale.Resolve(r),
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
