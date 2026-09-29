package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/order"
	"plantpal-backend/internal/interface/http/authmw"
	"plantpal-backend/internal/interface/http/reqlocale"
	"plantpal-backend/internal/interface/http/respond"
	orderuc "plantpal-backend/internal/usecase/order"
)

type OrderHandler struct {
	svc *orderuc.Service
}

func NewOrderHandler(svc *orderuc.Service) *OrderHandler {
	return &OrderHandler{svc: svc}
}

type createOrderItemRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type createOrderRequest struct {
	Items            []createOrderItemRequest `json:"items"`
	ShippingName     string                   `json:"shippingName"`
	ShippingPhone    string                   `json:"shippingPhone"`
	ShippingAddress  string                   `json:"shippingAddress"`
	DeliveryOptionID string                   `json:"deliveryOptionId"`
	PaymentMethodID  string                   `json:"paymentMethodId"`
}

// Create handles POST /api/v1/orders — "Continue to Payment" then a
// successful payment attempt on the Checkout/Payment screens. Prices the
// cart server-side from the live catalog and persists the order.
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, fmt.Errorf("%w: invalid JSON body", apperr.ErrInvalidInput))
		return
	}

	items := make([]orderuc.ItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, orderuc.ItemInput{ProductID: it.ProductID, Quantity: it.Quantity})
	}

	userID, _ := authmw.UserID(r.Context())
	o, err := h.svc.Create(r.Context(), orderuc.CreateInput{
		UserID:           userID,
		Items:            items,
		ShippingName:     req.ShippingName,
		ShippingPhone:    req.ShippingPhone,
		ShippingAddress:  req.ShippingAddress,
		DeliveryOptionID: req.DeliveryOptionID,
		PaymentMethodID:  req.PaymentMethodID,
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, o.Localized(reqlocale.Resolve(r)))
}

// List handles GET /api/v1/orders — order history, newest first.
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	lang := reqlocale.Resolve(r)
	localized := make([]order.Order, len(items))
	for i, item := range items {
		localized[i] = item.Localized(lang)
	}
	respond.JSON(w, http.StatusOK, localized)
}

// Get handles GET /api/v1/orders/{id}.
func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := authmw.UserID(r.Context())
	o, err := h.svc.Get(r.Context(), r.PathValue("id"), userID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, o.Localized(reqlocale.Resolve(r)))
}
