package events

import "time"

const (
	TopicOrderCreated        = "order-created"
	TopicOrderPaid           = "order-paid"
	TopicOrderCancelled      = "order-cancelled"
	TopicOrderStatusChanged  = "order-status-changed"
	TopicOrderTimeoutCheck   = "order-timeout-check"
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
