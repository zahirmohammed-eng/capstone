package model

import "time"

type OrderStatus string

const (
	StatusCreated   OrderStatus = "CREATED"
	StatusConfirmed OrderStatus = "CONFIRMED"
	StatusCancelled OrderStatus = "CANCELLED"
)

// Order GORM Model
type Order struct {
	ID          int64       `gorm:"primaryKey;autoIncrement" json:"id"`
	CustomerID  int64       `gorm:"index:idx_orders_customer_id;not null" json:"customerId"`
	ProductID   int64       `gorm:"not null" json:"productId"`
	Quantity    int         `gorm:"not null" json:"quantity"`
	UnitPrice   float64     `gorm:"type:numeric(12,2);not null" json:"unitPrice"`
	TotalAmount float64     `gorm:"type:numeric(12,2);not null" json:"totalAmount"`
	Status      OrderStatus `gorm:"type:varchar(50);index:idx_orders_status;not null" json:"status"`
	CreatedAt   time.Time   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
}

type CreateOrderRequest struct {
	CustomerID int64 `json:"customerId" binding:"required,gt=0"`
	ProductID  int64 `json:"productId" binding:"required,gt=0"`
	Quantity   int   `json:"quantity" binding:"required,gt=0"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}
