package models

import (
	"time"

	"github.com/google/uuid"
)

// Order represents an order in the system
type Order struct {
	ID         uuid.UUID   `json:"id"`
	UserID     uuid.UUID   `json:"user_id"`
	Status     OrderStatus `json:"status"`
	TotalPrice float64     `json:"total_price"`
	Items      []OrderItem `json:"items"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// OrderItem represents an item in an order
type OrderItem struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	Product   *Product  `json:"product,omitempty"`
}

// CreateOrderRequest represents the request payload for creating an order
type CreateOrderRequest struct {
	Items []CreateOrderItemRequest `json:"items" validate:"required,min=1"`
}

// CreateOrderItemRequest represents an item in a create order request
type CreateOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

// UpdateOrderStatusRequest represents the request payload for updating order status
type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" validate:"required,oneof=pending confirmed shipped delivered cancelled"`
}

// OrderResponse represents the response payload for order operations
type OrderResponse struct {
	ID         uuid.UUID         `json:"id"`
	UserID     uuid.UUID         `json:"user_id"`
	Status     OrderStatus       `json:"status"`
	TotalPrice float64           `json:"total_price"`
	Items      []OrderItemResponse `json:"items"`
	CreatedAt  time.Time         `json:"created_at"`
}

// OrderItemResponse represents an order item in responses
type OrderItemResponse struct {
	ID        uuid.UUID       `json:"id"`
	ProductID uuid.UUID       `json:"product_id"`
	Quantity  int             `json:"quantity"`
	UnitPrice float64         `json:"unit_price"`
	Product   ProductResponse `json:"product"`
}
