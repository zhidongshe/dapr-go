package models

import (
	"time"
)

// Inventory represents product stock information
type Inventory struct {
	ProductID      int64     `json:"product_id" db:"product_id"`
	ProductName    string    `json:"product_name" db:"product_name"`
	AvailableStock int       `json:"available_stock" db:"available_stock"`
	ReservedStock  int       `json:"reserved_stock" db:"reserved_stock"`
	Version        int       `json:"version" db:"version"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// InventoryReservation represents stock reservation for an order
type InventoryReservation struct {
	ID        int64     `json:"id" db:"id"`
	OrderNo   string    `json:"order_no" db:"order_no"`
	ProductID int64     `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Status    int       `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

const (
	ReservationStatusReserved  = 0
	ReservationStatusConfirmed = 1
	ReservationStatusReleased  = 2
)

// ReserveRequest represents a stock reservation request
type ReserveRequest struct {
	OrderNo string        `json:"order_no" binding:"required"`
	Items   []ReserveItem `json:"items" binding:"required,min=1"`
}

type ReserveItem struct {
	ProductID   int64  `json:"product_id" binding:"required"`
	ProductName string `json:"product_name" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required,gt=0"`
}

// ReleaseRequest represents a stock release request
type ReleaseRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
	Reason  string `json:"reason"`
}
