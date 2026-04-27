package models

import (
    "time"
    "fmt"
)

// Product status constants (mirrored from product-service)
const (
    ProductStatusOffSale = 0
    ProductStatusOnSale  = 1
)

// ProductSnapshot represents product data fetched from product service
type ProductSnapshot struct {
    ProductID     int64     `json:"product_id"`
    ProductName   string    `json:"product_name"`
    OriginalPrice int64     `json:"original_price"` // Price in cents
    Status        int       `json:"status"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

type Order struct {
    ID           uint64      `json:"id" db:"id"`
    OrderNo      string      `json:"order_no" db:"order_no"`
    UserID       uint64      `json:"user_id" db:"user_id"`
    TotalAmount  float64     `json:"total_amount" db:"total_amount"`
    Status       int         `json:"status" db:"status"`
    PayStatus    int         `json:"pay_status" db:"pay_status"`
    PayTime      *time.Time  `json:"pay_time,omitempty" db:"pay_time"`
    PayMethod    string      `json:"pay_method,omitempty" db:"pay_method"`
    Remark       string      `json:"remark,omitempty" db:"remark"`
    CreatedAt    time.Time   `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
    Items        []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
    ID          uint64  `json:"id" db:"id"`
    OrderID     uint64  `json:"order_id" db:"order_id"`
    ProductID   uint64  `json:"product_id" db:"product_id"`
    ProductName string  `json:"product_name" db:"product_name"`
    UnitPrice   float64 `json:"unit_price" db:"unit_price"`
    Quantity    int     `json:"quantity" db:"quantity"`
    TotalPrice  float64 `json:"total_price" db:"total_price"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type CreateOrderRequest struct {
    UserID uint64             `json:"user_id" binding:"required"`
    Items  []OrderItemRequest `json:"items" binding:"required,min=1"`
    Remark string             `json:"remark"`
}

type OrderItemRequest struct {
    ProductID uint64 `json:"product_id" binding:"required"`
    Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

type OrderResponse struct {
    OrderID     uint64    `json:"order_id"`
    OrderNo     string    `json:"order_no"`
    TotalAmount float64   `json:"total_amount"`
    Status      int       `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
}

func GenerateOrderNo() string {
    return fmt.Sprintf("ORD%s%04d",
        time.Now().Format("20060102150405"),
        time.Now().Nanosecond()%10000)
}
