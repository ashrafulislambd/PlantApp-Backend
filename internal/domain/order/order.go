// Package order holds the Order entity and its repository contract: a
// priced snapshot of the cart, shipping info, delivery choice, and payment
// method captured when a user completes checkout.
package order

import "time"

// Status is the lifecycle stage of an Order. Only StatusProcessing is
// produced today, matching the Flutter order-success screen, which always
// shows "Processing" immediately after a payment attempt succeeds. The
// type exists so later stages (confirmed, shipped, delivered, cancelled)
// are a non-breaking addition once a real payment/fulfilment flow exists.
type Status string

const StatusProcessing Status = "processing"

// Item is one line of an Order: a snapshot of a Product's name, image, and
// price at the moment of purchase, so later catalog changes never alter
// the record of a past order.
type Item struct {
	ProductID    string  `json:"productId" bson:"productId"`
	Name         string  `json:"name" bson:"name"`
	ImageURL     string  `json:"imageUrl,omitempty" bson:"imageUrl,omitempty"`
	UnitPriceBDT float64 `json:"unitPriceBdt" bson:"unitPriceBdt"`
	Quantity     int     `json:"quantity" bson:"quantity"`
	LineTotalBDT float64 `json:"lineTotalBdt" bson:"lineTotalBdt"`
}

// ShippingInfo is where the order is delivered, taken from the checkout
// form (name/phone/address fields on the Flutter Checkout screen).
type ShippingInfo struct {
	Name    string `json:"name" bson:"name"`
	Phone   string `json:"phone" bson:"phone"`
	Address string `json:"address" bson:"address"`
}

// DeliveryOption is one of the fixed shipping tiers the usecase resolves
// from a client-sent id (see usecase/order.deliveryOptions) — Title/ETA/fee
// are never trusted from the request itself.
//
// The Bn fields hold Bengali translations of Title/ETA; they're excluded
// from JSON directly (json:"-") and only surfaced through Localized. They
// still need bson tags (also "-"-style via a dedicated key) so they
// persist to Mongo alongside the rest of the order document.
type DeliveryOption struct {
	ID     string  `json:"id" bson:"id"`
	Title  string  `json:"title" bson:"title"`
	ETA    string  `json:"eta" bson:"eta"`
	FeeBDT float64 `json:"feeBdt" bson:"feeBdt"`

	TitleBn string `json:"-" bson:"titleBn,omitempty"`
	ETABn   string `json:"-" bson:"etaBn,omitempty"`
}

// Localized returns a copy with Title/ETA swapped for their Bengali
// translation when lang is "bn" and a translation exists.
func (d DeliveryOption) Localized(lang string) DeliveryOption {
	if lang != "bn" || d.TitleBn == "" {
		return d
	}
	d.Title = d.TitleBn
	if d.ETABn != "" {
		d.ETA = d.ETABn
	}
	return d
}

// Order is a completed checkout: cart contents plus shipping, delivery,
// and payment choice, priced server-side from the live product catalog.
type Order struct {
	ID              string         `json:"id" bson:"_id"`
	UserID          string         `json:"userId" bson:"userId"`
	Items           []Item         `json:"items" bson:"items"`
	Shipping        ShippingInfo   `json:"shipping" bson:"shipping"`
	Delivery        DeliveryOption `json:"delivery" bson:"delivery"`
	PaymentMethodID string         `json:"paymentMethodId" bson:"paymentMethodId"`
	SubtotalBDT     float64        `json:"subtotalBdt" bson:"subtotalBdt"`
	TaxBDT          float64        `json:"taxBdt" bson:"taxBdt"`
	TotalBDT        float64        `json:"totalBdt" bson:"totalBdt"`
	Status          Status         `json:"status" bson:"status"`
	CreatedAt       time.Time      `json:"createdAt" bson:"createdAt"`
}

// Localized returns a copy of the Order with its embedded DeliveryOption
// localized. Order itself has nothing else to translate: Items are a
// point-in-time snapshot of Product.Name, and Product isn't localized
// anywhere else in this codebase either.
func (o Order) Localized(lang string) Order {
	o.Delivery = o.Delivery.Localized(lang)
	return o
}
