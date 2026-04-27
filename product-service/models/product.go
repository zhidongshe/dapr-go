package models

import "time"

const (
	ProductStatusOffSale = 0
	ProductStatusOnSale  = 1
)

type Product struct {
	ProductID      int64     `json:"product_id" db:"product_id"`
	ProductName    string    `json:"product_name" db:"product_name"`
	OriginalPrice  int64     `json:"original_price" db:"original_price"`
	Status         int       `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type CreateProductRequest struct {
	ProductName   string `json:"product_name" binding:"required"`
	OriginalPrice int64  `json:"original_price" binding:"required,gt=0"`
	Status        *int   `json:"status,omitempty"`
}

type UpdatePriceRequest struct {
	OriginalPrice int64 `json:"original_price" binding:"required,gt=0"`
}

type UpdateStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1"`
}
