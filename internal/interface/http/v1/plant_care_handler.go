package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/plant"
	"plantpal-backend/internal/interface/http/authmw"
	"plantpal-backend/internal/interface/http/reqlocale"
	"plantpal-backend/internal/interface/http/reqtz"
	"plantpal-backend/internal/interface/http/respond"
	plantuc "plantpal-backend/internal/usecase/plant"
)

// updatePlantRequest's JSON keys match the Flutter client's PlantDto -
// notably `lastWatered`, which is how
// lib/features/plants/presentation/providers/plants_provider.dart's
// markWatered() actually marks a plant watered today (a PATCH, not the
// dedicated POST /plants/{id}/water endpoint below).
type updatePlantRequest struct {
	Name                  *string    `json:"name,omitempty"`
	Nickname              *string    `json:"nickname,omitempty"`
	Type                  *string    `json:"type,omitempty"`
	Species               *string    `json:"species,omitempty"`
	AgeStage              *string    `json:"ageStage,omitempty"`
	Location              *string    `json:"location,omitempty"`
	Sunlight              *string    `json:"sunlight,omitempty"`
	Outdoor               *bool      `json:"outdoor,omitempty"`
	Image                 *string    `json:"image,omitempty"`
	Status                *string    `json:"status,omitempty"`
	Humidity              *string    `json:"humidity,omitempty"`
	Health                *int       `json:"health,omitempty"`
	WateringFrequencyDays *int       `json:"wateringFrequencyDays,omitempty"`
	LastWateredAt         *time.Time `json:"lastWateredAt,omitempty"`
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

	name := req.Name
	if name == nil {
		name = req.Nickname
	}
	typ := req.Type
	if typ == nil {
		typ = req.Species
	}
	lastWatered := req.LastWateredAt
	if lastWatered == nil {
		lastWatered = req.LastWatered
	}

	userID, _ := authmw.UserID(r.Context())
	ctx := plant.ContextWithLocation(r.Context(), reqtz.Resolve(r))
	p, err := h.svc.Update(ctx, r.PathValue("id"), userID, plantuc.UpdateInput{
		Name:                  name,
		Type:                  typ,
		AgeStage:              req.AgeStage,
		Location:              req.Location,
		Sunlight:              req.Sunlight,
		Outdoor:               req.Outdoor,
		Image:                 req.Image,
		Status:                req.Status,
		Humidity:              req.Humidity,
		Health:                req.Health,
		WateringFrequencyDays: req.WateringFrequencyDays,
		LastWatered:           lastWatered,
		LastScan:              req.LastScan,
		Lang:                  reqlocale.Resolve(r),
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p.Localized(reqlocale.Resolve(r)))
}

// MarkWatered handles "I watered it".
func (h *PlantHandler) MarkWatered(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	ctx := plant.ContextWithLocation(r.Context(), reqtz.Resolve(r))
	result, err := h.svc.MarkWateredWithPoints(ctx, r.PathValue("id"), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{
		"plant":        result.Plant.Localized(reqlocale.Resolve(r)),
		"pointsEarned": result.PointsEarned,
	})
}

// MarkFertilized handles "I fertilized it".
func (h *PlantHandler) MarkFertilized(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	result, err := h.svc.MarkFertilizedWithPoints(r.Context(), r.PathValue("id"), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{
		"plant":        result.Plant.Localized(reqlocale.Resolve(r)),
		"pointsEarned": result.PointsEarned,
	})
}

type skipPlantRequest struct {
	Reason string `json:"reason"`
	Days   *int   `json:"days,omitempty"`
}

func (h *PlantHandler) Skip(w http.ResponseWriter, r *http.Request) {
	var req skipPlantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}
	days := 0
	if req.Days != nil {
		days = *req.Days
	}
	userID, _ := authmw.UserID(r.Context())
	p, err := h.svc.Skip(r.Context(), r.PathValue("id"), userID, req.Reason, days)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p.Localized(reqlocale.Resolve(r)))
}

type addPlantNoteRequest struct {
	Note string `json:"note"`
}

func (h *PlantHandler) AddNote(w http.ResponseWriter, r *http.Request) {
	var req addPlantNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}
	userID, _ := authmw.UserID(r.Context())
	if err := h.svc.AddNote(r.Context(), r.PathValue("id"), userID, req.Note); err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, map[string]string{"status": "recorded"})
}

func (h *PlantHandler) Events(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	var (
		events any
		err    error
	)
	if id := r.PathValue("id"); id != "" {
		events, err = h.svc.PlantEvents(r.Context(), id, userID)
	} else {
		events, err = h.svc.GardenEvents(r.Context(), userID)
	}
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, events)
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
	respond.JSON(w, http.StatusOK, plant.LocalizedAll(plants, reqlocale.Resolve(r)))
}
