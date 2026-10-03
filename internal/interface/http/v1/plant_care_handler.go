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

// updatePlantRequest's JSON keys match the Flutter client's PlantDto -
// notably `lastWatered`, which is how
// lib/features/plants/presentation/providers/plants_provider.dart's
// markWatered() actually marks a plant watered today (a PATCH, not the
// dedicated POST /plants/{id}/water endpoint below).
type updatePlantRequest struct {
	Name                  *string    `json:"nickname,omitempty"`
	Type                  *string    `json:"species,omitempty"`
	AgeStage              *string    `json:"ageStage,omitempty"`
	Location              *string    `json:"location,omitempty"`
	Sunlight              *string    `json:"sunlight,omitempty"`
	Image                 *string    `json:"image,omitempty"`
	Status                *string    `json:"status,omitempty"`
	Humidity              *string    `json:"humidity,omitempty"`
	Health                *int       `json:"health,omitempty"`
	WateringFrequencyDays *int       `json:"wateringFrequencyDays,omitempty"`
	LastWatered           *time.Time `json:"lastWatered,omitempty"`
	LastScan              *time.Time `json:"lastScan,omitempty"`
}

// Update handles editing any subset of a plant's fields from "My Plants".
func (h *PlantHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req updatePlantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}
	userID, _ := authmw.UserID(r.Context())
	p, err := h.svc.Update(r.Context(), r.PathValue("id"), userID, plantuc.UpdateInput{
		Name: req.Name, Type: req.Type, AgeStage: req.AgeStage,
		Location: req.Location, Sunlight: req.Sunlight, Image: req.Image,
		Status: req.Status, Humidity: req.Humidity, Health: req.Health,
		WateringFrequencyDays: req.WateringFrequencyDays,
		LastWatered:           req.LastWatered, LastScan: req.LastScan,
		Lang: reqlocale.Resolve(r),
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

// MarkWatered handles "I watered it".
func (h *PlantHandler) MarkWatered(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	p, err := h.svc.MarkWatered(r.Context(), r.PathValue("id"), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

// MarkFertilized handles "I fertilized it".
func (h *PlantHandler) MarkFertilized(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	p, err := h.svc.MarkFertilized(r.Context(), r.PathValue("id"), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

// Due lists the signed-in user's plants due for watering/fertilizing now,
// or by an optional ?before=<RFC3339> cutoff. Polled by the notifications
// module.
func (h *PlantHandler) Due(w http.ResponseWriter, r *http.Request) {
	before := time.Now().UTC()
	if q := r.URL.Query().Get("before"); q != "" {
		parsed, err := time.Parse(time.RFC3339, q)
		if err != nil {
			respond.Error(w, fmt.Errorf("%w: before must be RFC3339", apperr.ErrInvalidInput))
			return
		}
		before = parsed
	}
	userID, _ := authmw.UserID(r.Context())
	items, err := h.svc.Due(r.Context(), userID, before)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, items)
}
