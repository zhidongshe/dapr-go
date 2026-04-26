package repository

import (
	"testing"
	"time"

	"github.com/dapr-oms/inventory-service/models"
)

func TestInventoryRepositoryStruct(t *testing.T) {
	// Verify the repository struct can be instantiated
	repo := &InventoryRepository{}
	if repo == nil {
		t.Error("InventoryRepository should not be nil")
	}
}

func TestInventoryModel(t *testing.T) {
	inv := models.Inventory{
		ProductID:      1,
		ProductName:    "Test Product",
		AvailableStock: 100,
		ReservedStock:  10,
		Version:        1,
		UpdatedAt:      time.Now(),
	}

	if inv.ProductID != 1 {
		t.Error("ProductID mismatch")
	}

	if inv.AvailableStock < 0 {
		t.Error("AvailableStock should not be negative")
	}

	if inv.ReservedStock < 0 {
		t.Error("ReservedStock should not be negative")
	}
}

func TestInventoryReservationModel(t *testing.T) {
	reservation := models.InventoryReservation{
		ID:        1,
		OrderNo:   "ORD20250101120000",
		ProductID: 1,
		Quantity:  5,
		Status:    models.ReservationStatusReserved,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if reservation.OrderNo == "" {
		t.Error("OrderNo should not be empty")
	}

	if reservation.Quantity <= 0 {
		t.Error("Quantity should be positive")
	}

	if reservation.Status != models.ReservationStatusReserved {
		t.Errorf("Status should be %d (ReservationStatusReserved), got %d", models.ReservationStatusReserved, reservation.Status)
	}
}

func TestReservationStatusConstants(t *testing.T) {
	// Verify reservation status constants are properly defined
	if models.ReservationStatusReserved != 0 {
		t.Error("ReservationStatusReserved should be 0")
	}

	if models.ReservationStatusConfirmed != 1 {
		t.Error("ReservationStatusConfirmed should be 1")
	}

	if models.ReservationStatusReleased != 2 {
		t.Error("ReservationStatusReleased should be 2")
	}
}

func TestInventoryStockCalculation(t *testing.T) {
	// Test stock calculation logic
	inv := models.Inventory{
		AvailableStock: 100,
		ReservedStock:  20,
	}

	totalStock := inv.AvailableStock + inv.ReservedStock
	if totalStock != 120 {
		t.Errorf("Total stock should be 120, got %d", totalStock)
	}

	// Simulate reservation
	reserveQty := 10
	inv.AvailableStock -= reserveQty
	inv.ReservedStock += reserveQty

	if inv.AvailableStock != 90 {
		t.Errorf("AvailableStock should be 90, got %d", inv.AvailableStock)
	}

	if inv.ReservedStock != 30 {
		t.Errorf("ReservedStock should be 30, got %d", inv.ReservedStock)
	}
}
