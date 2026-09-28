package memory

import (
	"context"
	"testing"
	"time"

	"plantpal-backend/internal/domain/order"
)

func TestOrderRepository_List_ScopedToUserAndNewestFirst(t *testing.T) {
	repo := NewOrderRepository()
	ctx := context.Background()
	now := time.Now()

	mine1 := &order.Order{ID: "ord_1", UserID: "user_1", CreatedAt: now.Add(-time.Hour)}
	mine2 := &order.Order{ID: "ord_2", UserID: "user_1", CreatedAt: now}
	someoneElses := &order.Order{ID: "ord_3", UserID: "user_2", CreatedAt: now}

	for _, o := range []*order.Order{mine1, mine2, someoneElses} {
		if err := repo.Create(ctx, o); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	list, err := repo.List(ctx, "user_1")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("List() returned %d orders, want 2", len(list))
	}
	if list[0].ID != "ord_2" || list[1].ID != "ord_1" {
		t.Errorf("List() = [%s, %s], want [ord_2, ord_1] (newest first)", list[0].ID, list[1].ID)
	}
}

func TestOrderRepository_GetByID_NotFoundForOtherUser(t *testing.T) {
	repo := NewOrderRepository()
	ctx := context.Background()
	if err := repo.Create(ctx, &order.Order{ID: "ord_1", UserID: "user_1"}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := repo.GetByID(ctx, "ord_1", "user_2"); err == nil {
		t.Error("GetByID() error = nil, want not-found for a different user")
	}
	if _, err := repo.GetByID(ctx, "missing", "user_1"); err == nil {
		t.Error("GetByID() error = nil, want not-found for a missing id")
	}
}
