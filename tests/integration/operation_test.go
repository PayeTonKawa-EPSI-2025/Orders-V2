package integration_test

import (
	"context"
	"testing"

	"github.com/PayeTonKawa-EPSI-2025/Orders-V2/internal/operation"
)

func TestIntegration_GetOrders(t *testing.T) {
	db := ConnectDB(t)
	ResetOrdersTable(t, db)
	SeedDB(t, db)

	resp, err := operation.GetOrders(context.Background(), db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Body.Orders) != 2 {
		t.Fatalf("expected 2 customers, got %d", len(resp.Body.Orders))
	}
}
