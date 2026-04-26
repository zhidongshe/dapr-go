package models

import (
	"fmt"
	"time"
)

// PaymentStatus 支付状态常量
const (
	PaymentStatusPending  = 0 // 待支付
	PaymentStatusSuccess  = 1 // 支付成功
	PaymentStatusFailed   = 2 // 支付失败
	PaymentStatusRefunded = 3 // 已退款
)

// Payment 支付记录模型
type Payment struct {
	ID            int64      `json:"id" db:"id"`
	OrderNo       string     `json:"order_no" db:"order_no"`
	OrderID       int64      `json:"order_id" db:"order_id"`
	TransactionID string     `json:"transaction_id" db:"transaction_id"`
	Amount        float64    `json:"amount" db:"amount"`
	PayMethod     string     `json:"pay_method" db:"pay_method"`
	Status        int        `json:"status" db:"status"`
	PayTime       *time.Time `json:"pay_time,omitempty" db:"pay_time"`
	FailReason    string     `json:"fail_reason,omitempty" db:"fail_reason"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OrderNo   string  `json:"order_no" binding:"required"`
	PayMethod string  `json:"pay_method" binding:"required"`
	Amount    float64 `json:"amount"`
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message,omitempty"`
}

// PaymentCallbackRequest 支付回调请求
type PaymentCallbackRequest struct {
	OrderNo       string `json:"order_no" binding:"required"`
	TransactionID string `json:"transaction_id" binding:"required"`
	Status        string `json:"status" binding:"required"` // success/failed
}

// OrderInfo 订单信息
type OrderInfo struct {
	ID          int64   `json:"id"`
	OrderNo     string  `json:"order_no"`
	UserID      int64   `json:"user_id"`
	TotalAmount float64 `json:"total_amount"`
	Status      int     `json:"status"`
}

// GenerateTransactionID 生成交易号
func GenerateTransactionID() string {
	return fmt.Sprintf("TXN%s%04d",
		time.Now().Format("20060102150405"),
		time.Now().Nanosecond()%10000)
}

// GetStatusText 获取状态文本
func GetStatusText(status int) string {
	switch status {
	case PaymentStatusPending:
		return "pending"
	case PaymentStatusSuccess:
		return "success"
	case PaymentStatusFailed:
		return "failed"
	case PaymentStatusRefunded:
		return "refunded"
	default:
		return "unknown"
	}
}
