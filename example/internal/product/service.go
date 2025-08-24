package product

import (
	"fmt"

	"github.com/example/ecommerce-api/internal/models"
	"github.com/google/uuid"
)

// Service handles product business logic
type Service struct {
	repo *Repository
}

// ProvideService creates a new product service
func ProvideService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateProduct creates a new product
func (s *Service) CreateProduct(req *models.CreateProductRequest) (*models.ProductResponse, error) {
	// Validate business rules
	if err := s.validateCreateProductRequest(req); err != nil {
		return nil, err
	}

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
	}

	if err := s.repo.CreateProduct(product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return s.toProductResponse(product), nil
}

// GetProduct retrieves a product by ID
func (s *Service) GetProduct(id uuid.UUID) (*models.ProductResponse, error) {
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return s.toProductResponse(product), nil
}

// GetProducts retrieves all products
func (s *Service) GetProducts() ([]*models.ProductResponse, error) {
	products, err := s.repo.GetAllProducts()
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	responses := make([]*models.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = s.toProductResponse(product)
	}

	return responses, nil
}

// GetProductsByCategory retrieves products by category
func (s *Service) GetProductsByCategory(categoryID uuid.UUID) ([]*models.ProductResponse, error) {
	products, err := s.repo.GetProductsByCategory(categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products by category: %w", err)
	}

	responses := make([]*models.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = s.toProductResponse(product)
	}

	return responses, nil
}

// UpdateProduct updates a product
func (s *Service) UpdateProduct(id uuid.UUID, req *models.UpdateProductRequest) (*models.ProductResponse, error) {
	// Validate business rules
	if err := s.validateUpdateProductRequest(req); err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Stock != nil {
		updates["stock"] = *req.Stock
	}
	if req.CategoryID != nil {
		updates["category_id"] = *req.CategoryID
	}

	product, err := s.repo.UpdateProduct(id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return s.toProductResponse(product), nil
}

// DeleteProduct deletes a product
func (s *Service) DeleteProduct(id uuid.UUID) error {
	if err := s.repo.DeleteProduct(id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// GetCategories retrieves all categories
func (s *Service) GetCategories() ([]*models.Category, error) {
	categories, err := s.repo.GetAllCategories()
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

// CheckStock checks if enough stock is available
func (s *Service) CheckStock(id uuid.UUID, quantity int) (bool, error) {
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		return false, fmt.Errorf("failed to get product: %w", err)
	}

	return product.Stock >= quantity, nil
}

// ReserveStock reduces product stock (for order processing)
func (s *Service) ReserveStock(id uuid.UUID, quantity int) error {
	if err := s.repo.UpdateStock(id, -quantity); err != nil {
		return fmt.Errorf("failed to reserve stock: %w", err)
	}

	return nil
}

// ReleaseStock increases product stock (for order cancellation)
func (s *Service) ReleaseStock(id uuid.UUID, quantity int) error {
	if err := s.repo.UpdateStock(id, quantity); err != nil {
		return fmt.Errorf("failed to release stock: %w", err)
	}

	return nil
}

// validateCreateProductRequest validates create product request
func (s *Service) validateCreateProductRequest(req *models.CreateProductRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Name) < 2 || len(req.Name) > 100 {
		return fmt.Errorf("name must be between 2 and 100 characters")
	}
	if len(req.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}
	if req.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	if req.Stock < 0 {
		return fmt.Errorf("stock must be greater than or equal to 0")
	}
	return nil
}

// validateUpdateProductRequest validates update product request
func (s *Service) validateUpdateProductRequest(req *models.UpdateProductRequest) error {
	if req.Name != nil && (len(*req.Name) < 2 || len(*req.Name) > 100) {
		return fmt.Errorf("name must be between 2 and 100 characters")
	}
	if req.Description != nil && len(*req.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}
	if req.Price != nil && *req.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	if req.Stock != nil && *req.Stock < 0 {
		return fmt.Errorf("stock must be greater than or equal to 0")
	}
	return nil
}

// toProductResponse converts a Product model to ProductResponse
func (s *Service) toProductResponse(product *models.Product) *models.ProductResponse {
	return &models.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CategoryID:  product.CategoryID,
		CreatedAt:   product.CreatedAt,
	}
}
