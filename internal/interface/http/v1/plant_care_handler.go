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

type updatePlantRequest struct {
	Name                  *string `json:"name,omitempty"`
	Type                  *string `json:"type,omitempty"`
	AgeStage              *string `json:"ageStage,omitempty"`
	Location              *string `json:"location,omitempty"`
	Sunlight              *string `json:"sunlight,omitempty"`
	WateringFrequencyDays *int    `json:"wateringFrequencyDays,omitempty"`
}

// Update handles editing a plant's name/type/location/sunlight/watering-frequency from "My Plants".
func (h *PlantHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req updatePlantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}
	userID, _ := authmw.UserID(r.Context())
	p, err := h.svc.Update(r.Context(), r.PathValue("id"), userID, plantuc.UpdateInput{
		Name:                  req.Name,
		Type:                  req.Type,
		AgeStage:              req.AgeStage,
		Location:              req.Location,
		Sunlight:              req.Sunlight,
		WateringFrequencyDays: req.WateringFrequencyDays,
		Lang:                  reqlocale.Resolve(r),
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
	if raw := r.URL.Query().Get("before"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			respond.Error(w, fmt.Errorf("%w: ?before must be RFC3339", apperr.ErrInvalidInput))
			return
		}
		before = t
	}

	userID, _ := authmw.UserID(r.Context())
	plants, err := h.svc.Due(r.Context(), userID, before)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, plants)
}
