package models

import (
	"fmt"
	"time"
)

type CreatePaymentRequest struct {
	OrderNo   string  `json:"order_no" binding:"required"`
	PayMethod string  `json:"pay_method" binding:"required"`
	Amount    float64 `json:"amount"`
}

type PaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message,omitempty"`
}

type PaymentCallbackRequest struct {
	OrderNo       string `json:"order_no" binding:"required"`
	TransactionID string `json:"transaction_id" binding:"required"`
	Status        string `json:"status" binding:"required"` // success/failed
}

type OrderInfo struct {
	ID          int64   `json:"id"`
	OrderNo     string  `json:"order_no"`
	UserID      int64   `json:"user_id"`
	TotalAmount float64 `json:"total_amount"`
	Status      int     `json:"status"`
}

func GenerateTransactionID() string {
	return fmt.Sprintf("TXN%s%04d",
		time.Now().Format("20060102150405"),
		time.Now().Nanosecond()%10000)
}
