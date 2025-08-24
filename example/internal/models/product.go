package models

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a product in the system
type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CategoryID  uuid.UUID `json:"category_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequest represents the request payload for creating a product
type CreateProductRequest struct {
	Name        string    `json:"name" validate:"required,min=2,max=100"`
	Description string    `json:"description" validate:"max=500"`
	Price       float64   `json:"price" validate:"required,gt=0"`
	Stock       int       `json:"stock" validate:"gte=0"`
	CategoryID  uuid.UUID `json:"category_id" validate:"required"`
}

// UpdateProductRequest represents the request payload for updating a product
type UpdateProductRequest struct {
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=500"`
	Price       *float64   `json:"price,omitempty" validate:"omitempty,gt=0"`
	Stock       *int       `json:"stock,omitempty" validate:"omitempty,gte=0"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty"`
}

// ProductResponse represents the response payload for product operations
type ProductResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CategoryID  uuid.UUID `json:"category_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// Category represents a product category
type Category struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
