package v1

import (
	"net/http"

	"myplantpal-backend/internal/interface/http/respond"
	productuc "myplantpal-backend/internal/usecase/product"
)

type ProductHandler struct {
	svc *productuc.Service
}

func NewProductHandler(svc *productuc.Service) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// List handles GET /api/v1/products[?category=plants]
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	cat := r.URL.Query().Get("category")
	items, err := h.svc.List(cat)
	if err != nil {
		respond.Error(w, err)
		return
	}
	cats, _ := h.svc.ListCategories()
	respond.JSON(w, http.StatusOK, map[string]any{
		"categories": cats,
		"products":   items,
	})
}

// Get handles GET /api/v1/products/{id}
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.Get(r.PathValue("id"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, p)
}

// Refresh handles POST /api/v1/products/{id}/refresh
// Calls Groq Compound to get a live price, cached 6 h server-side.
func (h *ProductHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.RefreshPrice(r.PathValue("id"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, result)
}
