package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dapr-oms/inventory-service/models"
	_ "github.com/go-sql-driver/mysql"
)

type InventoryRepository struct {
	db *sql.DB
}

func NewInventoryRepository(dsn string) (*InventoryRepository, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &InventoryRepository{db: db}, nil
}

func (r *InventoryRepository) Close() error {
	return r.db.Close()
}

// GetInventory retrieves inventory by product ID
func (r *InventoryRepository) GetInventory(productID int64) (*models.Inventory, error) {
	inv := &models.Inventory{}
	err := r.db.QueryRow(
		`SELECT product_id, product_name, available_stock, reserved_stock, version, updated_at
		 FROM inventory WHERE product_id = ?`,
		productID,
	).Scan(&inv.ProductID, &inv.ProductName, &inv.AvailableStock, &inv.ReservedStock, &inv.Version, &inv.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// ReserveStock reserves stock for an order using optimistic locking
func (r *InventoryRepository) ReserveStock(tx *sql.Tx, productID int64, quantity int) error {
	result, err := tx.Exec(
		`UPDATE inventory
		 SET available_stock = available_stock - ?,
		     reserved_stock = reserved_stock + ?,
		     version = version + 1
		 WHERE product_id = ? AND available_stock >= ?`,
		quantity, quantity, productID, quantity,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("insufficient stock for product %d", productID)
	}
	return nil
}

// ConfirmStock confirms the reservation (moves from reserved to actual deduction)
func (r *InventoryRepository) ConfirmStock(tx *sql.Tx, productID int64, quantity int) error {
	result, err := tx.Exec(
		`UPDATE inventory
		 SET reserved_stock = reserved_stock - ?
		 WHERE product_id = ? AND reserved_stock >= ?`,
		quantity, productID, quantity,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("reservation not found for product %d", productID)
	}
	return nil
}

// ReleaseStock releases reserved stock back to available
func (r *InventoryRepository) ReleaseStock(tx *sql.Tx, productID int64, quantity int) error {
	result, err := tx.Exec(
		`UPDATE inventory
		 SET available_stock = available_stock + ?,
		     reserved_stock = reserved_stock - ?
		 WHERE product_id = ? AND reserved_stock >= ?`,
		quantity, quantity, productID, quantity,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("reservation not found for product %d", productID)
	}
	return nil
}

// CreateReservation creates a reservation record
func (r *InventoryRepository) CreateReservation(tx *sql.Tx, orderNo string, productID int64, quantity int) error {
	_, err := tx.Exec(
		`INSERT INTO inventory_reservation (order_no, product_id, quantity, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, NOW(), NOW())`,
		orderNo, productID, quantity, models.ReservationStatusReserved,
	)
	return err
}

// UpdateReservationStatus updates reservation status
func (r *InventoryRepository) UpdateReservationStatus(tx *sql.Tx, orderNo string, productID int64, status int) error {
	_, err := tx.Exec(
		`UPDATE inventory_reservation
		 SET status = ?, updated_at = NOW()
		 WHERE order_no = ? AND product_id = ?`,
		status, orderNo, productID,
	)
	return err
}

// GetReservationsByOrder retrieves all reservations for an order
func (r *InventoryRepository) GetReservationsByOrder(orderNo string) ([]models.InventoryReservation, error) {
	rows, err := r.db.Query(
		`SELECT id, order_no, product_id, quantity, status, created_at, updated_at
		 FROM inventory_reservation WHERE order_no = ?`,
		orderNo,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []models.InventoryReservation
	for rows.Next() {
		var r models.InventoryReservation
		err := rows.Scan(&r.ID, &r.OrderNo, &r.ProductID, &r.Quantity, &r.Status, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, r)
	}
	return reservations, nil
}

// BeginTransaction starts a new transaction
func (r *InventoryRepository) BeginTransaction() (*sql.Tx, error) {
	return r.db.Begin()
}
