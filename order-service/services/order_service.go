package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr-oms/order-service/models"
	"github.com/dapr-oms/order-service/repository"
	"github.com/dapr-oms/shared/events"
)

type OrderService struct {
	repo           *repository.OrderRepository
	daprClient     dapr.Client
	timeoutMinutes int
}

func NewOrderService() *OrderService {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(mysql:3306)/oms_db?charset=utf8mb4&parseTime=true"
	}

	repo, err := repository.NewOrderRepository(dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	client, err := dapr.NewClient()
	if err != nil {
		panic(fmt.Sprintf("failed to create dapr client: %v", err))
	}

	// 读取超时配置，默认10分钟
	timeoutMinutes := 10
	if timeoutStr := os.Getenv("ORDER_TIMEOUT_MINUTES"); timeoutStr != "" {
		if val, err := strconv.Atoi(timeoutStr); err == nil && val > 0 {
			timeoutMinutes = val
		}
	}

	return &OrderService{
		repo:           repo,
		daprClient:     client,
		timeoutMinutes: timeoutMinutes,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	var totalAmount float64
	items := make([]models.OrderItem, len(req.Items))
	for i, itemReq := range req.Items {
		items[i] = models.OrderItem{
			ProductID:   itemReq.ProductID,
			ProductName: itemReq.ProductName,
			UnitPrice:   itemReq.UnitPrice,
			Quantity:    itemReq.Quantity,
			TotalPrice:  float64(itemReq.Quantity) * itemReq.UnitPrice,
			CreatedAt:   time.Now(),
		}
		totalAmount += items[i].TotalPrice
	}

	order := &models.Order{
		OrderNo:     models.GenerateOrderNo(),
		UserID:      req.UserID,
		TotalAmount: totalAmount,
		Status:      events.OrderStatusPending,
		PayStatus:   events.PayStatusUnpaid,
		Remark:      req.Remark,
		Items:       items,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateOrder(order); err != nil {
		return nil, fmt.Errorf("create order failed: %w", err)
	}

	// Publish order created event
	event := events.OrderCreatedEvent{
		OrderID:     int64(order.ID),
		OrderNo:     order.OrderNo,
		UserID:      int64(order.UserID),
		TotalAmount: order.TotalAmount,
		Status:      order.Status,
		CreatedAt:   order.CreatedAt,
	}
	if err := s.publishEvent(ctx, events.TopicOrderCreated, event); err != nil {
		fmt.Printf("failed to publish order created event: %v\n", err)
	}

	// Publish order status changed event (pending is the initial status)
	statusEvent := events.OrderStatusChangedEvent{
		OrderID:   int64(order.ID),
		OrderNo:   order.OrderNo,
		UserID:    int64(order.UserID),
		OldStatus: -1, // -1 indicates initial status
		NewStatus: order.Status,
		ChangedAt: time.Now(),
	}
	if err := s.publishEvent(ctx, events.TopicOrderStatusChanged, statusEvent); err != nil {
		fmt.Printf("failed to publish status changed event: %v\n", err)
	}

	// Publish inventory reserve event
	inventoryEvent := events.InventoryReserveEvent{
		MessageID: generateUUID(),
		OrderID:   int64(order.ID),
		OrderNo:   order.OrderNo,
		UserID:    int64(order.UserID),
		Items:     convertToInventoryItems(order.Items),
		CreatedAt: time.Now(),
	}
	if err := s.publishEvent(ctx, events.TopicInventoryReserve, inventoryEvent); err != nil {
		fmt.Printf("failed to publish inventory reserve event: %v\n", err)
	}

	// Schedule timeout check using background goroutine
	go s.scheduleTimeoutCheck(order)

	return order, nil
}

// scheduleTimeoutCheck runs in background and cancels order after timeout
func (s *OrderService) scheduleTimeoutCheck(order *models.Order) {
	timeoutDuration := time.Duration(s.timeoutMinutes) * time.Minute
	fmt.Printf("scheduled timeout check for order %s in %d minutes\n", order.OrderNo, s.timeoutMinutes)

	time.Sleep(timeoutDuration)

	// Create a new context for background operation
	ctx := context.Background()

	// Reload order from database
	currentOrder, err := s.repo.GetOrderByID(order.ID)
	if err != nil {
		fmt.Printf("timeout check: get order failed: %v\n", err)
		return
	}
	if currentOrder == nil {
		fmt.Printf("timeout check: order not found: %d\n", order.ID)
		return
	}

	// Only cancel if still pending
	if currentOrder.Status != events.OrderStatusPending {
		fmt.Printf("timeout check: order %s is not pending (status=%d), skip cancel\n", currentOrder.OrderNo, currentOrder.Status)
		return
	}

	// Cancel the order
	fmt.Printf("timeout check: order %s timeout, auto cancelling...\n", currentOrder.OrderNo)
	if err := s.updateOrderStatusWithEvent(ctx, currentOrder, events.OrderStatusCancelled); err != nil {
		fmt.Printf("timeout check: cancel order failed: %v\n", err)
		return
	}

	// Publish order cancelled event
	cancelEvent := events.OrderCancelledEvent{
		OrderID:    int64(currentOrder.ID),
		OrderNo:    currentOrder.OrderNo,
		UserID:     int64(currentOrder.UserID),
		CancelTime: time.Now(),
		Reason:     "auto cancelled due to timeout",
	}
	if err := s.publishEvent(ctx, events.TopicOrderCancelled, cancelEvent); err != nil {
		fmt.Printf("timeout check: failed to publish cancelled event: %v\n", err)
	}

	fmt.Printf("timeout check: order %s auto cancelled successfully\n", currentOrder.OrderNo)
}

// updateOrderStatusWithEvent updates order status and publishes status changed event
func (s *OrderService) updateOrderStatusWithEvent(ctx context.Context, order *models.Order, newStatus int) error {
	oldStatus := order.Status
	if oldStatus == newStatus {
		return nil
	}

	if err := s.repo.UpdateOrderStatus(order.ID, newStatus); err != nil {
		return err
	}

	// Publish status changed event
	statusEvent := events.OrderStatusChangedEvent{
		OrderID:   int64(order.ID),
		OrderNo:   order.OrderNo,
		UserID:    int64(order.UserID),
		OldStatus: oldStatus,
		NewStatus: newStatus,
		ChangedAt: time.Now(),
	}
	if err := s.publishEvent(ctx, events.TopicOrderStatusChanged, statusEvent); err != nil {
		fmt.Printf("failed to publish status changed event: %v\n", err)
	}

	return nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID uint64) (*models.Order, error) {
	return s.repo.GetOrderByID(orderID)
}

func (s *OrderService) GetOrderByNo(ctx context.Context, orderNo string) (*models.Order, error) {
	return s.repo.GetOrderByNo(orderNo)
}

func (s *OrderService) ListOrders(ctx context.Context, userID uint64, limit, offset int) ([]models.Order, error) {
	return s.repo.ListOrders(userID, limit, offset)
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID uint64, reason string) error {
	order, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order not found")
	}

	if order.Status != events.OrderStatusPending {
		return fmt.Errorf("order cannot be cancelled, current status: %d", order.Status)
	}

	// Use unified status update method
	if err := s.updateOrderStatusWithEvent(ctx, order, events.OrderStatusCancelled); err != nil {
		return err
	}

	// Publish order cancelled event
	event := events.OrderCancelledEvent{
		OrderID:    int64(orderID),
		OrderNo:    order.OrderNo,
		UserID:     int64(order.UserID),
		CancelTime: time.Now(),
		Reason:     reason,
	}
	if err := s.publishEvent(ctx, events.TopicOrderCancelled, event); err != nil {
		fmt.Printf("failed to publish order cancelled event: %v\n", err)
	}

	// Publish inventory release event
	releaseEvent := events.InventoryReleaseEvent{
		MessageID:  generateUUID(),
		OrderID:    int64(order.ID),
		OrderNo:    order.OrderNo,
		Reason:     reason,
		ReleasedAt: time.Now(),
	}
	if err := s.publishEvent(ctx, events.TopicInventoryRelease, releaseEvent); err != nil {
		fmt.Printf("failed to publish inventory release event: %v\n", err)
	}

	return nil
}

func (s *OrderService) HandleOrderPaid(ctx context.Context, event *events.OrderPaidEvent) error {
	order, err := s.repo.GetOrderByID(uint64(event.OrderID))
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order not found: %d", event.OrderID)
	}

	// Update order status using unified method
	if err := s.updateOrderStatusWithEvent(ctx, order, events.OrderStatusPaid); err != nil {
		return err
	}

	// Update pay status
	if err := s.repo.UpdatePayStatus(order.ID, events.PayStatusPaid, event.PayTime, event.PayMethod); err != nil {
		return err
	}

	// Publish inventory confirm event
	confirmEvent := events.InventoryConfirmEvent{
		MessageID:   generateUUID(),
		OrderID:     event.OrderID,
		OrderNo:     event.OrderNo,
		ConfirmedAt: time.Now(),
	}
	if err := s.publishEvent(ctx, events.TopicInventoryConfirm, confirmEvent); err != nil {
		fmt.Printf("failed to publish inventory confirm event: %v\n", err)
	}

	return nil
}

func (s *OrderService) HandleInventoryReserveFailed(ctx context.Context, event *events.InventoryReserveFailedEvent) error {
	fmt.Printf("handling inventory reserve failed for order %s: %s\n", event.OrderNo, event.Reason)

	// Get order by order number
	order, err := s.repo.GetOrderByNo(event.OrderNo)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("order not found: %s", event.OrderNo)
	}

	// Only cancel if order is still pending
	if order.Status != events.OrderStatusPending {
		fmt.Printf("order %s is not pending (status=%d), skip cancel\n", event.OrderNo, order.Status)
		return nil
	}

	// Cancel the order
	if err := s.CancelOrder(ctx, order.ID, fmt.Sprintf("inventory reserve failed: %s", event.Reason)); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	fmt.Printf("order %s cancelled due to inventory reserve failed\n", event.OrderNo)
	return nil
}

func (s *OrderService) publishEvent(ctx context.Context, topic string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return s.daprClient.PublishEvent(ctx, "order-pubsub", topic, payload)
}

func convertToInventoryItems(items []models.OrderItem) []events.InventoryItem {
	result := make([]events.InventoryItem, len(items))
	for i, item := range items {
		result[i] = events.InventoryItem{
			ProductID:   int64(item.ProductID),
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
		}
	}
	return result
}

func generateUUID() string {
	// Simple UUID v4 generation - in production use proper UUID library
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}
