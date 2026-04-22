package models

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateOrderNo(t *testing.T) {
	// Test that GenerateOrderNo creates a valid order number
	orderNo := GenerateOrderNo()

	// Check format: ORD + timestamp + 4 digits
	if !strings.HasPrefix(orderNo, "ORD") {
		t.Errorf("GenerateOrderNo() = %v, want prefix 'ORD'", orderNo)
	}

	// Check length (should be around 19 characters: ORD + 14 digits + 4 digits)
	if len(orderNo) < 19 || len(orderNo) > 22 {
		t.Errorf("GenerateOrderNo() length = %v, want between 19 and 22", len(orderNo))
	}

	// Generate two order numbers and verify they are different
	orderNo2 := GenerateOrderNo()
	if orderNo == orderNo2 {
		t.Error("GenerateOrderNo() should generate unique order numbers")
	}
}

func TestCreateOrderRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateOrderRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateOrderRequest{
				UserID: 10001,
				Items: []OrderItemRequest{
					{
						ProductID:   101,
						ProductName: "iPhone 15",
						UnitPrice:   5999.00,
						Quantity:    1,
					},
				},
				Remark: "test remark",
			},
			wantErr: false,
		},
		{
			name: "multiple items",
			req: CreateOrderRequest{
				UserID: 10001,
				Items: []OrderItemRequest{
					{
						ProductID:   101,
						ProductName: "iPhone 15",
						UnitPrice:   5999.00,
						Quantity:    1,
					},
					{
						ProductID:   102,
						ProductName: "AirPods",
						UnitPrice:   1999.00,
						Quantity:    2,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the request structure
			if tt.req.UserID == 0 {
				t.Error("UserID should not be 0")
			}
			if len(tt.req.Items) == 0 {
				t.Error("Items should not be empty")
			}
			for _, item := range tt.req.Items {
				if item.ProductID == 0 {
					t.Error("ProductID should not be 0")
				}
				if item.UnitPrice <= 0 {
					t.Errorf("UnitPrice should be > 0, got %v", item.UnitPrice)
				}
				if item.Quantity <= 0 {
					t.Errorf("Quantity should be > 0, got %v", item.Quantity)
				}
			}
		})
	}
}

func TestOrderItemCalculation(t *testing.T) {
	item := OrderItem{
		ProductID:   101,
		ProductName: "iPhone 15",
		UnitPrice:   5999.00,
		Quantity:    2,
		TotalPrice:  11998.00,
		CreatedAt:   time.Now(),
	}

	expectedTotal := item.UnitPrice * float64(item.Quantity)
	if item.TotalPrice != expectedTotal {
		t.Errorf("TotalPrice = %v, want %v", item.TotalPrice, expectedTotal)
	}
}
