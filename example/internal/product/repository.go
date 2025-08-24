package product

import (
	"fmt"
	"sync"
	"time"

	"github.com/example/ecommerce-api/internal/models"
	"github.com/google/uuid"
)

// Repository handles product data persistence
type Repository struct {
	mu         sync.RWMutex
	products   map[uuid.UUID]*models.Product
	categories map[uuid.UUID]*models.Category
}

// ProvideRepository creates a new product repository
func ProvideRepository() *Repository {
	repo := &Repository{
		products:   make(map[uuid.UUID]*models.Product),
		categories: make(map[uuid.UUID]*models.Category),
	}

	// Add some default categories
	repo.seedCategories()
	return repo
}

// seedCategories adds some default categories
func (r *Repository) seedCategories() {
	categories := []*models.Category{
		{
			ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			Name:      "Electronics",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Name:      "Clothing",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
			Name:      "Books",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
			Name:      "Home & Garden",
			CreatedAt: time.Now(),
		},
	}

	for _, category := range categories {
		r.categories[category.ID] = category
	}
}

// CreateProduct creates a new product
func (r *Repository) CreateProduct(product *models.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if category exists
	if _, exists := r.categories[product.CategoryID]; !exists {
		return fmt.Errorf("category with ID %s not found", product.CategoryID)
	}

	product.ID = uuid.New()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	r.products[product.ID] = product
	return nil
}

// GetProductByID retrieves a product by ID
func (r *Repository) GetProductByID(id uuid.UUID) (*models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, exists := r.products[id]
	if !exists {
		return nil, fmt.Errorf("product with ID %s not found", id)
	}

	return product, nil
}

// GetAllProducts retrieves all products
func (r *Repository) GetAllProducts() ([]*models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := make([]*models.Product, 0, len(r.products))
	for _, product := range r.products {
		products = append(products, product)
	}

	return products, nil
}

// GetProductsByCategory retrieves products by category ID
func (r *Repository) GetProductsByCategory(categoryID uuid.UUID) ([]*models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var products []*models.Product
	for _, product := range r.products {
		if product.CategoryID == categoryID {
			products = append(products, product)
		}
	}

	return products, nil
}

// UpdateProduct updates a product
func (r *Repository) UpdateProduct(id uuid.UUID, updates map[string]interface{}) (*models.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	product, exists := r.products[id]
	if !exists {
		return nil, fmt.Errorf("product with ID %s not found", id)
	}

	// Create a copy to avoid modifying the original
	updatedProduct := *product

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		updatedProduct.Name = name
	}

	if description, ok := updates["description"].(string); ok {
		updatedProduct.Description = description
	}

	if price, ok := updates["price"].(float64); ok {
		updatedProduct.Price = price
	}

	if stock, ok := updates["stock"].(int); ok {
		updatedProduct.Stock = stock
	}

	if categoryID, ok := updates["category_id"].(uuid.UUID); ok {
		// Check if category exists
		if _, exists := r.categories[categoryID]; !exists {
			return nil, fmt.Errorf("category with ID %s not found", categoryID)
		}
		updatedProduct.CategoryID = categoryID
	}

	updatedProduct.UpdatedAt = time.Now()
	r.products[id] = &updatedProduct

	return &updatedProduct, nil
}

// DeleteProduct deletes a product
func (r *Repository) DeleteProduct(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.products[id]; !exists {
		return fmt.Errorf("product with ID %s not found", id)
	}

	delete(r.products, id)
	return nil
}

// GetAllCategories retrieves all categories
func (r *Repository) GetAllCategories() ([]*models.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make([]*models.Category, 0, len(r.categories))
	for _, category := range r.categories {
		categories = append(categories, category)
	}

	return categories, nil
}

// GetCategoryByID retrieves a category by ID
func (r *Repository) GetCategoryByID(id uuid.UUID) (*models.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	category, exists := r.categories[id]
	if !exists {
		return nil, fmt.Errorf("category with ID %s not found", id)
	}

	return category, nil
}

// UpdateStock updates product stock
func (r *Repository) UpdateStock(id uuid.UUID, quantity int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	product, exists := r.products[id]
	if !exists {
		return fmt.Errorf("product with ID %s not found", id)
	}

	if product.Stock+quantity < 0 {
		return fmt.Errorf("insufficient stock for product %s", id)
	}

	product.Stock += quantity
	product.UpdatedAt = time.Now()

	return nil
}
