package order

import (
	"context"
	"errors"
	"math"
	"testing"

	"plantpal-backend/internal/domain/apperr"
	"plantpal-backend/internal/domain/product"
	"plantpal-backend/internal/idgen"
	"plantpal-backend/internal/infrastructure/repository/memory"
)

func newTestService() *Service {
	products := memory.NewProductRepository(
		[]*product.Product{
			{ID: "p1", Name: "Jade Plant", PriceBDT: 100},
			{ID: "p2", Name: "Watering Can", PriceBDT: 50},
		},
		nil,
	)
	return NewService(memory.NewOrderRepository(), products, idgen.New())
}

func validInput() CreateInput {
	return CreateInput{
		UserID:           "user_1",
		Items:            []ItemInput{{ProductID: "p1", Quantity: 2}, {ProductID: "p2", Quantity: 1}},
		ShippingName:     "Promitee",
		ShippingPhone:    "01700000000",
		ShippingAddress:  "House 1, Road 2, Dhaka",
		DeliveryOptionID: "standard",
		PaymentMethodID:  "cod",
	}
}

func almostEqual(a, b float64) bool { return math.Abs(a-b) < 0.01 }

func TestCreate_RequiresUserID(t *testing.T) {
	svc := newTestService()
	in := validInput()
	in.UserID = ""
	if _, err := svc.Create(context.Background(), in); !errors.Is(err, apperr.ErrInvalidInput) {
		t.Errorf("Create() error = %v, want %v", err, apperr.ErrInvalidInput)
	}
}

func TestCreate_RequiresNonEmptyCart(t *testing.T) {
	svc := newTestService()
	in := validInput()
	in.Items = nil
	if _, err := svc.Create(context.Background(), in); !errors.Is(err, apperr.ErrInvalidInput) {
		t.Errorf("Create() error = %v, want %v", err, apperr.ErrInvalidInput)
	}
}

func TestCreate_RequiresShippingFields(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(in *CreateInput)
	}{
		{"missing name", func(in *CreateInput) { in.ShippingName = "  " }},
		{"short phone", func(in *CreateInput) { in.ShippingPhone = "123" }},
		{"missing address", func(in *CreateInput) { in.ShippingAddress = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTestService()
			in := validInput()
			tc.mutate(&in)
			if _, err := svc.Create(context.Background(), in); !errors.Is(err, apperr.ErrInvalidInput) {
				t.Errorf("Create() error = %v, want %v", err, apperr.ErrInvalidInput)
			}
		})
	}
}

func TestCreate_RejectsUnknownDeliveryOption(t *testing.T) {
	svc := newTestService()
	in := validInput()
	in.DeliveryOptionID = "overnight"
	if _, err := svc.Create(context.Background(), in); !errors.Is(err, apperr.ErrInvalidInput) {
		t.Errorf("Create() error = %v, want %v", err, apperr.ErrInvalidInput)
	}
}

func TestCreate_RejectsUnknownPaymentMethod(t *testing.T) {
	svc := newTestService()
	in := validInput()
	in.PaymentMethodID = "crypto"
	if _, err := svc.Create(context.Background(), in); !errors.Is(err, apperr.ErrInvalidInput) {
		t.Errorf("Create() error = %v, want %v", err, apperr.ErrInvalidInput)
	}
}

func TestCreate_RejectsUnknownProduct(t *testing.T) {
	svc := newTestService()
	in := validInput()
	in.Items = []ItemInput{{ProductID: "does-not-exist", Quantity: 1}}
	if _, err := svc.Create(context.Background(), in); !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("Create() error = %v, want %v", err, apperr.ErrNotFound)
	}
}

// TestCreate_PricesFromCatalog is the important one: it proves totals come
// from the seeded product prices (100*2 + 50*1 = 250 subtotal), never from
// anything the client could supply — ItemInput has no price field at all.
func TestCreate_PricesFromCatalog(t *testing.T) {
	svc := newTestService()
	o, err := svc.Create(context.Background(), validInput())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !almostEqual(o.SubtotalBDT, 250) {
		t.Errorf("SubtotalBDT = %v, want 250", o.SubtotalBDT)
	}
	if !almostEqual(o.TaxBDT, 12.5) {
		t.Errorf("TaxBDT = %v, want 12.5 (5%% of 250)", o.TaxBDT)
	}
	if !almostEqual(o.Delivery.FeeBDT, 60) {
		t.Errorf("Delivery.FeeBDT = %v, want 60 for standard", o.Delivery.FeeBDT)
	}
	if !almostEqual(o.TotalBDT, 322.5) {
		t.Errorf("TotalBDT = %v, want 322.5 (250 + 60 + 12.5)", o.TotalBDT)
	}
	if o.Status != "processing" {
		t.Errorf("Status = %v, want processing", o.Status)
	}
}

func TestGet_NotFoundForOtherUser(t *testing.T) {
	svc := newTestService()
	o, err := svc.Create(context.Background(), validInput())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := svc.Get(context.Background(), o.ID, "someone_else"); !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("Get() error = %v, want %v", err, apperr.ErrNotFound)
	}
}

func TestList_ReturnsCreatedOrders(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	if _, err := svc.Create(ctx, validInput()); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := svc.Create(ctx, validInput()); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	list, err := svc.List(ctx, "user_1")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List() returned %d orders, want 2", len(list))
	}
}
