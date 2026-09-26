package v1

import (
	"net/http"

	"myplantpal-backend/internal/domain/product"
	"myplantpal-backend/internal/interface/http/reqlocale"
	"myplantpal-backend/internal/interface/http/respond"
	productuc "myplantpal-backend/internal/usecase/product"
)


var categoryNamesBn = map[string]string{
	"plants":             "ইনডোর গাছ",
	"outdoor-plants":     "আউটডোর গাছ",
	"flowering-plants":   "ফুল গাছ",
	"succulents-cacti":   "সাকুলেন্ট ও ক্যাকটাস",
	"herbs-vegetables":   "ভেষজ ও শাকসবজি",
	"seeds":              "বীজ",
	"pots-planters":      "টব ও প্লান্টার",
	"soil-potting-mix":   "মাটি ও পটিং মিক্স",
	"fertilizer":         "সার",
	"care":               "গাছের যত্ন ও সুরক্ষা",
	"gardening-tools":    "বাগানের যন্ত্রপাতি",
	"watering-equipment": "পানি দেওয়ার সরঞ্জাম",
	"grow-lights":        "গ্রো লাইট ও সরঞ্জাম",
}

type ProductHandler struct {
	svc *productuc.Service
}

func NewProductHandler(svc *productuc.Service) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// List handles GET /api/v1/products[?categoryId=plants]
// The legacy ?category= parameter is still accepted; categoryId wins if both
// are present. The response always carries both categories and products.
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	cat := q.Get("categoryId")
	if cat == "" {
		cat = q.Get("category")
	}
	items, err := h.svc.List(cat)
	if err != nil {
		respond.Error(w, err)
		return
	}
	cats, err := h.svc.ListCategories()
	if err != nil {
		respond.Error(w, err)
		return
	}
	if reqlocale.Resolve(r) == reqlocale.BN {
		localizedCats := make([]*product.Category, len(cats))
		for i, c := range cats {
			copyCat := *c
			if bnName, ok := categoryNamesBn[c.ID]; ok {
				copyCat.Name = bnName
			}
			localizedCats[i] = &copyCat
		}
		cats = localizedCats
	}
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
