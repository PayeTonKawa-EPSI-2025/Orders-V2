package operation_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PayeTonKawa-EPSI-2025/Orders-V2/internal/operation"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbMock,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm DB: %v", err)
	}

	return gormDB, mock
}

func TestGetOrders(t *testing.T) {
	db, mock := setupMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "customer_id"}).
		AddRow(1, "1").
		AddRow(2, "3")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders"`)).WillReturnRows(rows)

	resp, err := operation.GetOrders(context.Background(), db)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Body.Orders) != 2 {
		t.Errorf("expected 2 orders, got %d", len(resp.Body.Orders))
	}

	if resp.Body.Orders[0].ID != 1 {
		t.Errorf("expected first order '1', got '%d'", resp.Body.Orders[0].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled sqlmock expectations: %v", err)
	}
}

func TestGetOrderNotFound(t *testing.T) {
	db, mock := setupMockDB(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" WHERE "orders"."id" = $1`)).
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := operation.GetOrder(context.Background(), db, 1)
	if err == nil {
		t.Fatal("expected error for non-existent order")
	}
}
