package events

import (
	"testing"
	"time"
)

func TestOrderStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected int
	}{
		{"OrderStatusPending", OrderStatusPending, 0},
		{"OrderStatusPaid", OrderStatusPaid, 1},
		{"OrderStatusProcessing", OrderStatusProcessing, 2},
		{"OrderStatusShipped", OrderStatusShipped, 3},
		{"OrderStatusCompleted", OrderStatusCompleted, 4},
		{"OrderStatusCancelled", OrderStatusCancelled, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.status, tt.expected)
			}
		})
	}
}

func TestPayStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected int
	}{
		{"PayStatusUnpaid", PayStatusUnpaid, 0},
		{"PayStatusPaid", PayStatusPaid, 1},
		{"PayStatusFailed", PayStatusFailed, 2},
		{"PayStatusRefunded", PayStatusRefunded, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.status, tt.expected)
			}
		})
	}
}

func TestOrderCreatedEvent(t *testing.T) {
	event := OrderCreatedEvent{
		OrderID:     1,
		OrderNo:     "ORD202504220001",
		UserID:      10001,
		TotalAmount: 5999.00,
		Status:      OrderStatusPending,
		CreatedAt:   time.Now(),
	}

	if event.OrderID != 1 {
		t.Errorf("OrderID = %v, want 1", event.OrderID)
	}
	if event.Status != OrderStatusPending {
		t.Errorf("Status = %v, want OrderStatusPending", event.Status)
	}
}

func TestOrderPaidEvent(t *testing.T) {
	now := time.Now()
	event := OrderPaidEvent{
		OrderID:   1,
		OrderNo:   "ORD202504220001",
		UserID:    10001,
		OldStatus: OrderStatusPending,
		NewStatus: OrderStatusPaid,
		PayTime:   now,
		PayMethod: "alipay",
	}

	if event.OldStatus != OrderStatusPending {
		t.Errorf("OldStatus = %v, want OrderStatusPending", event.OldStatus)
	}
	if event.NewStatus != OrderStatusPaid {
		t.Errorf("NewStatus = %v, want OrderStatusPaid", event.NewStatus)
	}
	if event.PayMethod != "alipay" {
		t.Errorf("PayMethod = %v, want alipay", event.PayMethod)
	}
}

func TestTopicConstants(t *testing.T) {
	if TopicOrderCreated != "order-created" {
		t.Errorf("TopicOrderCreated = %v, want order-created", TopicOrderCreated)
	}
	if TopicOrderPaid != "order-paid" {
		t.Errorf("TopicOrderPaid = %v, want order-paid", TopicOrderPaid)
	}
	if TopicOrderCancelled != "order-cancelled" {
		t.Errorf("TopicOrderCancelled = %v, want order-cancelled", TopicOrderCancelled)
	}
	if TopicOrderStatusChanged != "order-status-changed" {
		t.Errorf("TopicOrderStatusChanged = %v, want order-status-changed", TopicOrderStatusChanged)
	}
}
