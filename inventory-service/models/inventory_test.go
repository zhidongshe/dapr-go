package models

import (
	"testing"
	"time"
)

func TestInventoryConstants(t *testing.T) {
	if ReservationStatusReserved != 0 {
		t.Errorf("ReservationStatusReserved should be 0")
	}
	if ReservationStatusConfirmed != 1 {
		t.Errorf("ReservationStatusConfirmed should be 1")
	}
	if ReservationStatusReleased != 2 {
		t.Errorf("ReservationStatusReleased should be 2")
	}
}

func TestInventoryStruct(t *testing.T) {
	inv := Inventory{
		ProductID:      1,
		ProductName:    "Test Product",
		AvailableStock: 100,
		ReservedStock:  10,
		Version:        1,
		UpdatedAt:      time.Now(),
	}

	if inv.ProductID != 1 {
		t.Errorf("ProductID mismatch")
	}
	if inv.ProductName != "Test Product" {
		t.Errorf("ProductName mismatch")
	}
}
