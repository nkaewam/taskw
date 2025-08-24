package api

import (
	"github.com/example/ecommerce-api/internal/models"
	"github.com/example/ecommerce-api/internal/order"
	"github.com/example/ecommerce-api/internal/product"
	"github.com/example/ecommerce-api/internal/user"
	"github.com/google/uuid"
)

// ProductServiceAdapter adapts product.Service to order.ProductService interface
type ProductServiceAdapter struct {
	service *product.Service
}

// ProvideProductServiceAdapter creates a new adapter
func ProvideProductServiceAdapter(service *product.Service) order.ProductService {
	return &ProductServiceAdapter{service: service}
}

func (a *ProductServiceAdapter) GetProduct(id uuid.UUID) (*models.ProductResponse, error) {
	return a.service.GetProduct(id)
}

func (a *ProductServiceAdapter) CheckStock(id uuid.UUID, quantity int) (bool, error) {
	return a.service.CheckStock(id, quantity)
}

func (a *ProductServiceAdapter) ReserveStock(id uuid.UUID, quantity int) error {
	return a.service.ReserveStock(id, quantity)
}

func (a *ProductServiceAdapter) ReleaseStock(id uuid.UUID, quantity int) error {
	return a.service.ReleaseStock(id, quantity)
}

// UserServiceAdapter adapts user.Service to order.UserService interface
type UserServiceAdapter struct {
	service *user.Service
}

// ProvideUserServiceAdapter creates a new adapter
func ProvideUserServiceAdapter(service *user.Service) order.UserService {
	return &UserServiceAdapter{service: service}
}

func (a *UserServiceAdapter) GetUser(id uuid.UUID) (*models.UserResponse, error) {
	return a.service.GetUser(id)
}
