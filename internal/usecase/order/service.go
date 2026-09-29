// Package order implements the checkout usecase: turns a cart + shipping +
// delivery + payment choice into a priced, persisted Order.
package order

import (
	"context"
	"fmt"
	"strings"
	"time"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/order"
	"plantpal-backend/internal/domain/product"
	"plantpal-backend/internal/idgen"
)

// deliveryOptions are the fixed shipping tiers a client may choose by id.
// Title/ETA/fee are resolved here and never trusted from the client, so a
// request can't under-report shipping cost. Mirrors DeliveryOption.all in
// the Flutter app's lib/features/checkout/domain/entities/checkout_models.dart
// — keep both in sync if either changes.
//
// TitleBn/ETABn are surfaced through order.Order.Localized, same as every
// other translated field in this codebase (see fertilizer.Fertilizer,
// diagnosis.Diagnosis).
var deliveryOptions = map[string]order.DeliveryOption{
	"standard": {
		ID: "standard", Title: "Standard Delivery", ETA: "3-5 Days", FeeBDT: 60,
		TitleBn: "স্ট্যান্ডার্ড ডেলিভারি", ETABn: "৩-৫ দিন",
	},
	"express": {
		ID: "express", Title: "Express Delivery", ETA: "1-2 Days", FeeBDT: 120,
		TitleBn: "এক্সপ্রেস ডেলিভারি", ETABn: "১-২ দিন",
	},
}

// paymentMethodIDs are the ids accepted from PaymentMethod.all in the
// Flutter app's lib/features/payments/domain/entities/payment_models.dart.
// Their display titles are already localized client-side, so the backend
// only needs to validate the id.
var paymentMethodIDs = map[string]bool{
	"card":  true,
	"bkash": true,
	"nagad": true,
	"cod":   true,
}

// taxRate matches CalculateOrderSummary.taxRate in checkout_models.dart.
const taxRate = 0.05

// Service is the checkout usecase.
type Service struct {
	repo     order.Repository
	products product.Repository
	ids      idgen.Generator
}

func NewService(repo order.Repository, products product.Repository, ids idgen.Generator) *Service {
	return &Service{repo: repo, products: products, ids: ids}
}

// ItemInput is one requested cart line: a product id and quantity. Price
// is never taken from the client — it's looked up from the catalog.
type ItemInput struct {
	ProductID string
	Quantity  int
}

type CreateInput struct {
	UserID           string
	Items            []ItemInput
	ShippingName     string
	ShippingPhone    string
	ShippingAddress  string
	DeliveryOptionID string
	PaymentMethodID  string
}

// Create prices Items against the live product catalog, validates shipping
// and the delivery/payment choice, and persists the result.
//
// There is no real payment gateway wired up yet, so payment itself isn't
// processed here — any request that passes validation is recorded as
// StatusProcessing, same as the Flutter app's current simulated-payment
// flow. Swap in a real charge behind this same signature once a payment
// provider exists.
func (s *Service) Create(ctx context.Context, in CreateInput) (*order.Order, error) {
	if strings.TrimSpace(in.UserID) == "" {
		return nil, fmt.Errorf("%w: userID is required", apperr.ErrInvalidInput)
	}
	if len(in.Items) == 0 {
		return nil, fmt.Errorf("%w: cart is empty", apperr.ErrInvalidInput)
	}

	name := strings.TrimSpace(in.ShippingName)
	phone := strings.TrimSpace(in.ShippingPhone)
	address := strings.TrimSpace(in.ShippingAddress)
	if name == "" {
		return nil, fmt.Errorf("%w: shipping name is required", apperr.ErrInvalidInput)
	}
	if len(phone) < 7 {
		return nil, fmt.Errorf("%w: a valid shipping phone number is required", apperr.ErrInvalidInput)
	}
	if address == "" {
		return nil, fmt.Errorf("%w: shipping address is required", apperr.ErrInvalidInput)
	}

	delivery, ok := deliveryOptions[strings.ToLower(strings.TrimSpace(in.DeliveryOptionID))]
	if !ok {
		return nil, fmt.Errorf("%w: unknown delivery option %q", apperr.ErrInvalidInput, in.DeliveryOptionID)
	}

	paymentMethodID := strings.ToLower(strings.TrimSpace(in.PaymentMethodID))
	if !paymentMethodIDs[paymentMethodID] {
		return nil, fmt.Errorf("%w: unknown payment method %q", apperr.ErrInvalidInput, in.PaymentMethodID)
	}

	items := make([]order.Item, 0, len(in.Items))
	var subtotal float64
	for _, it := range in.Items {
		if it.Quantity < 1 {
			return nil, fmt.Errorf("%w: quantity must be at least 1", apperr.ErrInvalidInput)
		}
		p, err := s.products.GetByID(it.ProductID)
		if err != nil {
			return nil, err
		}
		lineTotal := p.PriceBDT * float64(it.Quantity)
		items = append(items, order.Item{
			ProductID:    p.ID,
			Name:         p.Name,
			ImageURL:     p.ImageURL,
			UnitPriceBDT: p.PriceBDT,
			Quantity:     it.Quantity,
			LineTotalBDT: lineTotal,
		})
		subtotal += lineTotal
	}

	tax := subtotal * taxRate
	total := subtotal + delivery.FeeBDT + tax

	now := time.Now().UTC()
	o := &order.Order{
		ID:     s.ids.New("ord"),
		UserID: in.UserID,
		Items:  items,
		Shipping: order.ShippingInfo{
			Name:    name,
			Phone:   phone,
			Address: address,
		},
		Delivery:        delivery,
		PaymentMethodID: paymentMethodID,
		SubtotalBDT:     subtotal,
		TaxBDT:          tax,
		TotalBDT:        total,
		Status:          order.StatusProcessing,
		CreatedAt:       now,
	}
	if err := s.repo.Create(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *Service) Get(ctx context.Context, id, userID string) (*order.Order, error) {
	return s.repo.GetByID(ctx, id, userID)
}

// List returns userID's order history, newest first.
func (s *Service) List(ctx context.Context, userID string) ([]*order.Order, error) {
	return s.repo.List(ctx, userID)
}
