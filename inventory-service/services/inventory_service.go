package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr-oms/inventory-service/models"
	"github.com/dapr-oms/inventory-service/repository"
	"github.com/dapr-oms/shared/events"
)

type InventoryService struct {
	repo       *repository.InventoryRepository
	daprClient dapr.Client
}

func NewInventoryService() *InventoryService {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(mysql:3306)/oms_db?charset=utf8mb4&parseTime=true"
	}

	repo, err := repository.NewInventoryRepository(dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	client, err := dapr.NewClient()
	if err != nil {
		panic(fmt.Sprintf("failed to create dapr client: %v", err))
	}

	return &InventoryService{
		repo:       repo,
		daprClient: client,
	}
}

// GetInventory retrieves inventory for a product
func (s *InventoryService) GetInventory(productID int64) (*models.Inventory, error) {
	return s.repo.GetInventory(productID)
}

// ReserveStock reserves stock for an order
func (s *InventoryService) ReserveStock(ctx context.Context, req *events.InventoryReserveEvent) error {
	tx, err := s.repo.BeginTransaction()
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Reserve stock for each item
	for _, item := range req.Items {
		if err := s.repo.ReserveStock(tx, item.ProductID, item.Quantity); err != nil {
			return fmt.Errorf("reserve stock failed for product %d: %w", item.ProductID, err)
		}

		// Create reservation record
		if err := s.repo.CreateReservation(tx, req.OrderNo, item.ProductID, item.Quantity); err != nil {
			return fmt.Errorf("create reservation record failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	// Publish success event
	event := events.InventoryReservedEvent{
		MessageID:  req.MessageID,
		OrderID:    req.OrderID,
		OrderNo:    req.OrderNo,
		ReservedAt: req.CreatedAt,
	}
	if err := s.publishEvent(ctx, events.TopicInventoryReserved, event); err != nil {
		fmt.Printf("failed to publish inventory reserved event: %v\n", err)
	}

	return nil
}

// ConfirmStock confirms the stock deduction after payment
func (s *InventoryService) ConfirmStock(ctx context.Context, req *events.InventoryConfirmEvent) error {
	// Get reservations for this order
	reservations, err := s.repo.GetReservationsByOrder(req.OrderNo)
	if err != nil {
		return fmt.Errorf("get reservations failed: %w", err)
	}

	if len(reservations) == 0 {
		return fmt.Errorf("no reservations found for order %s", req.OrderNo)
	}

	tx, err := s.repo.BeginTransaction()
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Confirm each reservation
	for _, r := range reservations {
		if r.Status != models.ReservationStatusReserved {
			continue // Already confirmed or released
		}

		if err := s.repo.ConfirmStock(tx, r.ProductID, r.Quantity); err != nil {
			return fmt.Errorf("confirm stock failed for product %d: %w", r.ProductID, err)
		}

		if err := s.repo.UpdateReservationStatus(tx, req.OrderNo, r.ProductID, models.ReservationStatusConfirmed); err != nil {
			return fmt.Errorf("update reservation status failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	// Publish success event
	event := events.InventoryConfirmedEvent{
		MessageID:   req.MessageID,
		OrderID:     req.OrderID,
		OrderNo:     req.OrderNo,
		ConfirmedAt: req.ConfirmedAt,
	}
	if err := s.publishEvent(ctx, events.TopicInventoryConfirmed, event); err != nil {
		fmt.Printf("failed to publish inventory confirmed event: %v\n", err)
	}

	return nil
}

// ReleaseStock releases reserved stock
func (s *InventoryService) ReleaseStock(ctx context.Context, req *events.InventoryReleaseEvent) error {
	// Get reservations for this order
	reservations, err := s.repo.GetReservationsByOrder(req.OrderNo)
	if err != nil {
		return fmt.Errorf("get reservations failed: %w", err)
	}

	if len(reservations) == 0 {
		fmt.Printf("no reservations to release for order %s\n", req.OrderNo)
		return nil
	}

	tx, err := s.repo.BeginTransaction()
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Release each reservation
	for _, r := range reservations {
		if r.Status != models.ReservationStatusReserved {
			continue // Already confirmed or released
		}

		if err := s.repo.ReleaseStock(tx, r.ProductID, r.Quantity); err != nil {
			return fmt.Errorf("release stock failed for product %d: %w", r.ProductID, err)
		}

		if err := s.repo.UpdateReservationStatus(tx, req.OrderNo, r.ProductID, models.ReservationStatusReleased); err != nil {
			return fmt.Errorf("update reservation status failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	// Publish success event
	event := events.InventoryReleasedEvent{
		MessageID:  req.MessageID,
		OrderID:    req.OrderID,
		OrderNo:    req.OrderNo,
		ReleasedAt: req.ReleasedAt,
	}
	if err := s.publishEvent(ctx, events.TopicInventoryReleased, event); err != nil {
		fmt.Printf("failed to publish inventory released event: %v\n", err)
	}

	return nil
}

// HandleReserveFailed handles reserve failure by publishing event
func (s *InventoryService) HandleReserveFailed(ctx context.Context, req *events.InventoryReserveEvent, reason string) {
	event := events.InventoryReserveFailedEvent{
		MessageID: req.MessageID,
		OrderID:   req.OrderID,
		OrderNo:   req.OrderNo,
		Reason:    reason,
		FailedAt:  req.CreatedAt,
	}
	if err := s.publishEvent(ctx, events.TopicInventoryReserveFailed, event); err != nil {
		fmt.Printf("failed to publish reserve failed event: %v\n", err)
	}
}

func (s *InventoryService) publishEvent(ctx context.Context, topic string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return s.daprClient.PublishEvent(ctx, "order-pubsub", topic, payload)
}
