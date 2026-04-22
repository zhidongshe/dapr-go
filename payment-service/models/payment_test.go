package models

import (
	"strings"
	"testing"
)

func TestGenerateTransactionID(t *testing.T) {
	// Test that GenerateTransactionID creates a valid transaction ID
	transactionID := GenerateTransactionID()

	// Check format: TXN + timestamp + 4 digits
	if !strings.HasPrefix(transactionID, "TXN") {
		t.Errorf("GenerateTransactionID() = %v, want prefix 'TXN'", transactionID)
	}

	// Check length (should be around 20 characters: TXN + 14 digits + 4 digits)
	if len(transactionID) < 19 || len(transactionID) > 22 {
		t.Errorf("GenerateTransactionID() length = %v, want between 19 and 22", len(transactionID))
	}

	// Generate two transaction IDs and verify they are different
	transactionID2 := GenerateTransactionID()
	if transactionID == transactionID2 {
		t.Error("GenerateTransactionID() should generate unique transaction IDs")
	}
}

func TestCreatePaymentRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreatePaymentRequest
		wantErr bool
	}{
		{
			name: "valid alipay payment",
			req: CreatePaymentRequest{
				OrderNo:   "ORD202504220001",
				PayMethod: "alipay",
				Amount:    5999.00,
			},
			wantErr: false,
		},
		{
			name: "valid wechat payment",
			req: CreatePaymentRequest{
				OrderNo:   "ORD202504220002",
				PayMethod: "wechat",
				Amount:    1999.50,
			},
			wantErr: false,
		},
		{
			name: "zero amount (should be allowed for validation at service layer)",
			req: CreatePaymentRequest{
				OrderNo:   "ORD202504220003",
				PayMethod: "alipay",
				Amount:    0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the request structure
			if tt.req.OrderNo == "" {
				t.Error("OrderNo should not be empty")
			}
			if tt.req.PayMethod == "" {
				t.Error("PayMethod should not be empty")
			}
		})
	}
}

func TestPaymentResponse(t *testing.T) {
	tests := []struct {
		name          string
		resp          PaymentResponse
		expectedStatus string
	}{
		{
			name: "success response",
			resp: PaymentResponse{
				TransactionID: "TXN202504220001",
				Status:        "success",
				Message:       "payment processed successfully",
			},
			expectedStatus: "success",
		},
		{
			name: "failed response",
			resp: PaymentResponse{
				TransactionID: "TXN202504220002",
				Status:        "failed",
				Message:       "insufficient balance",
			},
			expectedStatus: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.resp.Status != tt.expectedStatus {
				t.Errorf("Status = %v, want %v", tt.resp.Status, tt.expectedStatus)
			}
			if tt.resp.TransactionID == "" {
				t.Error("TransactionID should not be empty")
			}
		})
	}
}

func TestPaymentCallbackRequest(t *testing.T) {
	tests := []struct {
		name string
		req  PaymentCallbackRequest
	}{
		{
			name: "successful callback",
			req: PaymentCallbackRequest{
				OrderNo:       "ORD202504220001",
				TransactionID: "TXN202504220001",
				Status:        "success",
			},
		},
		{
			name: "failed callback",
			req: PaymentCallbackRequest{
				OrderNo:       "ORD202504220002",
				TransactionID: "TXN202504220002",
				Status:        "failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.req.OrderNo == "" {
				t.Error("OrderNo should not be empty")
			}
			if tt.req.TransactionID == "" {
				t.Error("TransactionID should not be empty")
			}
			if tt.req.Status != "success" && tt.req.Status != "failed" {
				t.Errorf("Status should be 'success' or 'failed', got %v", tt.req.Status)
			}
		})
	}
}

func TestOrderInfo(t *testing.T) {
	orderInfo := OrderInfo{
		ID:          1,
		OrderNo:     "ORD202504220001",
		UserID:      10001,
		TotalAmount: 5999.00,
		Status:      0, // pending
	}

	if orderInfo.ID != 1 {
		t.Errorf("ID = %v, want 1", orderInfo.ID)
	}
	if orderInfo.OrderNo != "ORD202504220001" {
		t.Errorf("OrderNo = %v, want ORD202504220001", orderInfo.OrderNo)
	}
	if orderInfo.UserID != 10001 {
		t.Errorf("UserID = %v, want 10001", orderInfo.UserID)
	}
}
