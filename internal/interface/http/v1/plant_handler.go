package v1

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/plant"
	"plantpal-backend/internal/interface/http/authmw"
	"plantpal-backend/internal/interface/http/reqlocale"
	"plantpal-backend/internal/interface/http/reqtz"
	"plantpal-backend/internal/interface/http/respond"
	plantuc "plantpal-backend/internal/usecase/plant"
)

type PlantHandler struct {
	svc           *plantuc.Service
	maxImageBytes int64
}

func NewPlantHandler(svc *plantuc.Service, maxImageBytes int64) *PlantHandler {
	if maxImageBytes <= 0 {
		maxImageBytes = 8 << 20
	}
	return &PlantHandler{svc: svc, maxImageBytes: maxImageBytes}
}

// createPlantRequest accepts both standard keys (name, type, lastWateredAt)
// and Flutter client keys (nickname, species, lastWatered).
type createPlantRequest struct {
	Name                  string     `json:"name"`
	Nickname              string     `json:"nickname"`
	Type                  string     `json:"type"`
	Species               string     `json:"species"`
	AgeStage              string     `json:"ageStage"`
	Location              string     `json:"location"`
	Sunlight              string     `json:"sunlight"`
	Outdoor               bool       `json:"outdoor"`
	Image                 string     `json:"image"`
	WateringFrequencyDays int        `json:"wateringFrequencyDays"`
	// LastWateredAt (RFC3339) is optional; omitted means the plant is due now.
	LastWateredAt *time.Time `json:"lastWateredAt,omitempty"`
	LastWatered   *time.Time `json:"lastWatered,omitempty"`
	// WaterAmountMl and CareTips are optional AI-identify values stored in
	// the care roadmap.
	WaterAmountMl int    `json:"waterAmountMl,omitempty"`
	CareTips      string `json:"careTips,omitempty"`
}

func (h *PlantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createPlantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = strings.TrimSpace(req.Nickname)
	}
	typ := strings.TrimSpace(req.Type)
	if typ == "" {
		typ = strings.TrimSpace(req.Species)
	}
	lastWatered := req.LastWateredAt
	if lastWatered == nil {
		lastWatered = req.LastWatered
	}

	userID, _ := authmw.UserID(r.Context())
	ctx := plant.ContextWithLocation(r.Context(), reqtz.Resolve(r))
	p, err := h.svc.Create(ctx, plantuc.CreateInput{
		UserID:                userID,
		Name:                  name,
		Type:                  typ,
		AgeStage:              req.AgeStage,
		Location:              req.Location,
		Sunlight:              req.Sunlight,
		Outdoor:               req.Outdoor,
		Image:                 req.Image,
		WateringFrequencyDays: req.WateringFrequencyDays,
		LastWateredAt:         lastWatered,
		WaterAmountMl:         req.WaterAmountMl,
		CareTips:              req.CareTips,
		Lang:                  reqlocale.Resolve(r),
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, p.Localized(reqlocale.Resolve(r)))
}

func (h *PlantHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, plant.LocalizedAll(items, reqlocale.Resolve(r)))
}

func (h *PlantHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	p, err := h.svc.Get(r.Context(), r.PathValue("id"), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p.Localized(reqlocale.Resolve(r)))
}

func (h *PlantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	if err := h.svc.Delete(r.Context(), r.PathValue("id"), userID); err != nil {
		respond.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type identifyRequest struct {
	ImageBase64 string `json:"imageBase64"`
}

func (h *PlantHandler) readImageBytes(w http.ResponseWriter, r *http.Request) ([]byte, string, error) {
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		r.Body = http.MaxBytesReader(w, r.Body, h.maxImageBytes+(64<<10))
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			return nil, "", fmt.Errorf("%w: invalid multipart form", apperr.ErrInvalidInput)
		}
		defer r.MultipartForm.RemoveAll()
		file, _, err := r.FormFile("image")
		if err != nil {
			return nil, "", fmt.Errorf("%w: multipart field \"image\" is required", apperr.ErrInvalidInput)
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, h.maxImageBytes+1))
		if err != nil {
			return nil, "", fmt.Errorf("read image: %w", err)
		}
		detectedCT := http.DetectContentType(data)
		return data, detectedCT, nil
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxImageBytes*4/3+4096)
	var req identifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, "", fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput)
	}
	data, err := base64.StdEncoding.DecodeString(req.ImageBase64)
	if err != nil {
		return nil, "", fmt.Errorf("%w: imageBase64 is not valid base64", apperr.ErrInvalidInput)
	}
	return data, "image/jpeg", nil
}

func (h *PlantHandler) Identify(w http.ResponseWriter, r *http.Request) {
	data, _, err := h.readImageBytes(w, r)
	if err != nil {
		respond.Error(w, err)
		return
	}
	res, err := h.svc.Identify(r.Context(), data)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, res)
}

func (h *PlantHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	plantID := r.PathValue("id")
	data, ct, err := h.readImageBytes(w, r)
	if err != nil {
		respond.Error(w, err)
		return
	}
	p, err := h.svc.UploadImage(r.Context(), plantID, userID, data, ct)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p.Localized(reqlocale.Resolve(r)))
}

func (h *PlantHandler) Image(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	plantID := r.PathValue("id")
	rc, ct, err := h.svc.Image(r.Context(), plantID, userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	_, _ = io.Copy(w, rc)
}
