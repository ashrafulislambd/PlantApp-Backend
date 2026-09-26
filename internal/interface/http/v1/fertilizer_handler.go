package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"myplantpal-backend/internal/domain/apperr"
	"myplantpal-backend/internal/domain/fertilizer"
	"myplantpal-backend/internal/interface/http/reqlocale"
	"myplantpal-backend/internal/interface/http/respond"
	fertilizeruc "myplantpal-backend/internal/usecase/fertilizer"
)

type FertilizerHandler struct {
	svc *fertilizeruc.Service
}

func NewFertilizerHandler(svc *fertilizeruc.Service) *FertilizerHandler {
	return &FertilizerHandler{svc: svc}
}

type createFertilizerRequest struct {
	Name         string `json:"name"`
	Category     string `json:"category"`
	Instructions string `json:"instructions"`
}

// Create handles "Add a new Fertilizer" on the Fertilizer screen.
func (h *FertilizerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createFertilizerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}

	f, err := h.svc.Create(r.Context(), fertilizeruc.CreateInput{
		Name:         req.Name,
		Category:     req.Category,
		Instructions: req.Instructions,
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, f.Localized(reqlocale.Resolve(r)))
}

// List handles the "Find your homemade fertilizer" search bar via ?q=.
func (h *FertilizerHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	lang := reqlocale.Resolve(r)
	localized := make([]fertilizer.Fertilizer, len(items))
	for i, item := range items {
		localized[i] = item.Localized(lang)
	}
	respond.JSON(w, http.StatusOK, localized)
}

func (h *FertilizerHandler) Get(w http.ResponseWriter, r *http.Request) {
	f, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, f.Localized(reqlocale.Resolve(r)))
}
