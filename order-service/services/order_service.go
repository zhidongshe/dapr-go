package services

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "time"

    dapr "github.com/dapr/go-sdk/client"
    "github.com/dapr-oms/order-service/models"
    "github.com/dapr-oms/order-service/repository"
    "github.com/dapr-oms/shared/events"
)

type OrderService struct {
    repo   *repository.OrderRepository
    daprClient dapr.Client
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

    return &OrderService{
        repo:   repo,
        daprClient: client,
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

    return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID uint64) (*models.Order, error) {
    return s.repo.GetOrderByID(orderID)
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

    if err := s.repo.UpdateOrderStatus(orderID, events.OrderStatusCancelled); err != nil {
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

    // Update order status
    if err := s.repo.UpdateOrderStatus(order.ID, events.OrderStatusPaid); err != nil {
        return err
    }

    // Update pay status
    if err := s.repo.UpdatePayStatus(order.ID, events.PayStatusPaid, event.PayTime, event.PayMethod); err != nil {
        return err
    }

    // Publish status changed event
    statusEvent := events.OrderStatusChangedEvent{
        OrderID:   event.OrderID,
        OrderNo:   event.OrderNo,
        UserID:    event.UserID,
        OldStatus: event.OldStatus,
        NewStatus: event.NewStatus,
        ChangedAt: time.Now(),
    }
    if err := s.publishEvent(ctx, events.TopicOrderStatusChanged, statusEvent); err != nil {
        fmt.Printf("failed to publish status changed event: %v\n", err)
    }

    return nil
}

func (s *OrderService) publishEvent(ctx context.Context, topic string, data interface{}) error {
    payload, err := json.Marshal(data)
    if err != nil {
        return err
    }

    return s.daprClient.PublishEvent(ctx, "order-pubsub", topic, payload)
}
