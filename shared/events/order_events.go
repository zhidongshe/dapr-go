package events

import (
	"encoding/json"
	"time"
)

const (
	TopicOrderCreated        = "order-created"
	TopicOrderPaid           = "order-paid"
	TopicOrderCancelled      = "order-cancelled"
	TopicOrderStatusChanged  = "order-status-changed"
	TopicOrderTimeoutCheck   = "order-timeout-check"
)

const (
	TopicInventoryReserve       = "inventory-reserve"
	TopicInventoryReserved      = "inventory-reserved"
	TopicInventoryReserveFailed = "inventory-reserve-failed"
	TopicInventoryConfirm       = "inventory-confirm"
	TopicInventoryConfirmed     = "inventory-confirmed"
	TopicInventoryRelease       = "inventory-release"
	TopicInventoryReleased      = "inventory-released"
	TopicDeadLetter             = "dead-letter"
)

const (
	OrderStatusPending    = 0
	OrderStatusPaid       = 1
	OrderStatusProcessing = 2
	OrderStatusShipped    = 3
	OrderStatusCompleted  = 4
	OrderStatusCancelled  = 5
)

const (
	PayStatusUnpaid   = 0
	PayStatusPaid     = 1
	PayStatusFailed   = 2
	PayStatusRefunded = 3
)

type OrderCreatedEvent struct {
	OrderID     int64     `json:"order_id"`
	OrderNo     string    `json:"order_no"`
	UserID      int64     `json:"user_id"`
	TotalAmount float64   `json:"total_amount"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type OrderPaidEvent struct {
	OrderID   int64     `json:"order_id"`
	OrderNo   string    `json:"order_no"`
	UserID    int64     `json:"user_id"`
	OldStatus int       `json:"old_status"`
	NewStatus int       `json:"new_status"`
	PayTime   time.Time `json:"pay_time"`
	PayMethod string    `json:"pay_method"`
}

type OrderCancelledEvent struct {
	OrderID    int64     `json:"order_id"`
	OrderNo    string    `json:"order_no"`
	UserID     int64     `json:"user_id"`
	CancelTime time.Time `json:"cancel_time"`
	Reason     string    `json:"reason,omitempty"`
}

type OrderStatusChangedEvent struct {
	OrderID   int64     `json:"order_id"`
	OrderNo   string    `json:"order_no"`
	UserID    int64     `json:"user_id"`
	OldStatus int       `json:"old_status"`
	NewStatus int       `json:"new_status"`
	ChangedAt time.Time `json:"changed_at"`
}

type OrderTimeoutCheckEvent struct {
	OrderID   int64     `json:"order_id"`
	OrderNo   string    `json:"order_no"`
	CreatedAt time.Time `json:"created_at"`
	CheckAt   time.Time `json:"check_at"`
}

// InventoryItem represents an item in inventory operations
type InventoryItem struct {
	ProductID   int64  `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

// InventoryReserveEvent - Step 1: Order Service publishes to reserve stock
type InventoryReserveEvent struct {
	MessageID string          `json:"message_id"`
	OrderID   int64           `json:"order_id"`
	OrderNo   string          `json:"order_no"`
	UserID    int64           `json:"user_id"`
	Items     []InventoryItem `json:"items"`
	CreatedAt time.Time       `json:"created_at"`
}

// InventoryReservedEvent - Step 2: Inventory Service publishes success
type InventoryReservedEvent struct {
	MessageID  string    `json:"message_id"`
	OrderID    int64     `json:"order_id"`
	OrderNo    string    `json:"order_no"`
	ReservedAt time.Time `json:"reserved_at"`
}

// InventoryReserveFailedEvent - Step 2 (alt): Inventory Service publishes failure
type InventoryReserveFailedEvent struct {
	MessageID string    `json:"message_id"`
	OrderID   int64     `json:"order_id"`
	OrderNo   string    `json:"order_no"`
	Reason    string    `json:"reason"`
	FailedAt  time.Time `json:"failed_at"`
}

// InventoryConfirmEvent - Step 3: Order Service publishes to confirm deduction
type InventoryConfirmEvent struct {
	MessageID   string    `json:"message_id"`
	OrderID     int64     `json:"order_id"`
	OrderNo     string    `json:"order_no"`
	ConfirmedAt time.Time `json:"confirmed_at"`
}

// InventoryConfirmedEvent - Step 4: Inventory Service publishes confirmation success
type InventoryConfirmedEvent struct {
	MessageID   string    `json:"message_id"`
	OrderID     int64     `json:"order_id"`
	OrderNo     string    `json:"order_no"`
	ConfirmedAt time.Time `json:"confirmed_at"`
}

// InventoryReleaseEvent - Alternative: Release reserved stock
type InventoryReleaseEvent struct {
	MessageID  string    `json:"message_id"`
	OrderID    int64     `json:"order_id"`
	OrderNo    string    `json:"order_no"`
	Reason     string    `json:"reason"`
	ReleasedAt time.Time `json:"released_at"`
}

// InventoryReleasedEvent - Confirmation of release
type InventoryReleasedEvent struct {
	MessageID  string    `json:"message_id"`
	OrderID    int64     `json:"order_id"`
	OrderNo    string    `json:"order_no"`
	ReleasedAt time.Time `json:"released_at"`
}

// DeadLetterMessage - Failed messages for manual handling
type DeadLetterMessage struct {
	OriginalTopic string          `json:"original_topic"`
	MessageID     string          `json:"message_id"`
	Payload       json.RawMessage `json:"payload"`
	Error         string          `json:"error"`
	FailedCount   int             `json:"failed_count"`
	CreatedAt     time.Time       `json:"created_at"`
}
