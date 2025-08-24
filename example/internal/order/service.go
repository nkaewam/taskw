package order

import (
	"fmt"

	"github.com/example/ecommerce-api/internal/models"
	"github.com/google/uuid"
)

// ProductService interface for interacting with product service
type ProductService interface {
	GetProduct(id uuid.UUID) (*models.ProductResponse, error)
	CheckStock(id uuid.UUID, quantity int) (bool, error)
	ReserveStock(id uuid.UUID, quantity int) error
	ReleaseStock(id uuid.UUID, quantity int) error
}

// UserService interface for interacting with user service
type UserService interface {
	GetUser(id uuid.UUID) (*models.UserResponse, error)
}

// Service handles order business logic
type Service struct {
	repo           *Repository
	productService ProductService
	userService    UserService
}

// ProvideService creates a new order service
func ProvideService(repo *Repository, productService ProductService, userService UserService) *Service {
	return &Service{
		repo:           repo,
		productService: productService,
		userService:    userService,
	}
}

// CreateOrder creates a new order
func (s *Service) CreateOrder(userID uuid.UUID, req *models.CreateOrderRequest) (*models.OrderResponse, error) {
	// Validate user exists
	if _, err := s.userService.GetUser(userID); err != nil {
		return nil, fmt.Errorf("invalid user: %w", err)
	}

	// Validate business rules
	if err := s.validateCreateOrderRequest(req); err != nil {
		return nil, err
	}

	// Calculate total price and validate stock
	var totalPrice float64
	var orderItems []models.OrderItem

	for _, item := range req.Items {
		// Get product details
		product, err := s.productService.GetProduct(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product %s: %w", item.ProductID, err)
		}

		// Check stock availability
		available, err := s.productService.CheckStock(item.ProductID, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to check stock for product %s: %w", item.ProductID, err)
		}

		if !available {
			return nil, fmt.Errorf("insufficient stock for product %s", product.Name)
		}

		// Create order item
		orderItem := models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: product.Price,
		}

		orderItems = append(orderItems, orderItem)
		totalPrice += product.Price * float64(item.Quantity)
	}

	// Create order
	order := &models.Order{
		UserID:     userID,
		Status:     models.OrderStatusPending,
		TotalPrice: totalPrice,
		Items:      orderItems,
	}

	if err := s.repo.CreateOrder(order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Reserve stock for all items
	for _, item := range order.Items {
		if err := s.productService.ReserveStock(item.ProductID, item.Quantity); err != nil {
			// Rollback: release already reserved stock and delete order
			s.rollbackStockReservation(order.Items, item.ProductID)
			s.repo.DeleteOrder(order.ID)
			return nil, fmt.Errorf("failed to reserve stock for product %s: %w", item.ProductID, err)
		}
	}

	return s.toOrderResponse(order), nil
}

// GetOrder retrieves an order by ID
func (s *Service) GetOrder(id uuid.UUID) (*models.OrderResponse, error) {
	order, err := s.repo.GetOrderByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return s.toOrderResponse(order), nil
}

// GetOrdersByUser retrieves all orders for a user
func (s *Service) GetOrdersByUser(userID uuid.UUID) ([]*models.OrderResponse, error) {
	// Validate user exists
	if _, err := s.userService.GetUser(userID); err != nil {
		return nil, fmt.Errorf("invalid user: %w", err)
	}

	orders, err := s.repo.GetOrdersByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders for user: %w", err)
	}

	responses := make([]*models.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = s.toOrderResponse(order)
	}

	return responses, nil
}

// GetAllOrders retrieves all orders
func (s *Service) GetAllOrders() ([]*models.OrderResponse, error) {
	orders, err := s.repo.GetAllOrders()
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	responses := make([]*models.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = s.toOrderResponse(order)
	}

	return responses, nil
}

// UpdateOrderStatus updates the status of an order
func (s *Service) UpdateOrderStatus(id uuid.UUID, req *models.UpdateOrderStatusRequest) (*models.OrderResponse, error) {
	// Validate status transition
	currentOrder, err := s.repo.GetOrderByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if err := s.validateStatusTransition(currentOrder.Status, req.Status); err != nil {
		return nil, err
	}

	// Handle stock release if order is cancelled
	if req.Status == models.OrderStatusCancelled && currentOrder.Status != models.OrderStatusCancelled {
		for _, item := range currentOrder.Items {
			if err := s.productService.ReleaseStock(item.ProductID, item.Quantity); err != nil {
				return nil, fmt.Errorf("failed to release stock for product %s: %w", item.ProductID, err)
			}
		}
	}

	order, err := s.repo.UpdateOrderStatus(id, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	return s.toOrderResponse(order), nil
}

// CancelOrder cancels an order
func (s *Service) CancelOrder(id uuid.UUID) error {
	req := &models.UpdateOrderStatusRequest{
		Status: models.OrderStatusCancelled,
	}

	_, err := s.UpdateOrderStatus(id, req)
	return err
}

// GetOrdersByStatus retrieves orders by status
func (s *Service) GetOrdersByStatus(status models.OrderStatus) ([]*models.OrderResponse, error) {
	orders, err := s.repo.GetOrdersByStatus(status)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status: %w", err)
	}

	responses := make([]*models.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = s.toOrderResponse(order)
	}

	return responses, nil
}

// validateCreateOrderRequest validates create order request
func (s *Service) validateCreateOrderRequest(req *models.CreateOrderRequest) error {
	if len(req.Items) == 0 {
		return fmt.Errorf("order must contain at least one item")
	}

	for i, item := range req.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be greater than 0", i)
		}
	}

	return nil
}

// validateStatusTransition validates order status transitions
func (s *Service) validateStatusTransition(current, new models.OrderStatus) error {
	validTransitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderStatusPending:   {models.OrderStatusConfirmed, models.OrderStatusCancelled},
		models.OrderStatusConfirmed: {models.OrderStatusShipped, models.OrderStatusCancelled},
		models.OrderStatusShipped:   {models.OrderStatusDelivered},
		models.OrderStatusDelivered: {}, // Final state
		models.OrderStatusCancelled: {}, // Final state
	}

	allowed := validTransitions[current]
	for _, status := range allowed {
		if status == new {
			return nil
		}
	}

	return fmt.Errorf("invalid status transition from %s to %s", current, new)
}

// rollbackStockReservation releases stock for items up to the failed product
func (s *Service) rollbackStockReservation(items []models.OrderItem, failedProductID uuid.UUID) {
	for _, item := range items {
		if item.ProductID == failedProductID {
			break
		}
		s.productService.ReleaseStock(item.ProductID, item.Quantity)
	}
}

// toOrderResponse converts an Order model to OrderResponse
func (s *Service) toOrderResponse(order *models.Order) *models.OrderResponse {
	items := make([]models.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		// Try to get product details for the response
		var productResponse models.ProductResponse
		if product, err := s.productService.GetProduct(item.ProductID); err == nil {
			productResponse = *product
		} else {
			// Fallback if product service is unavailable
			productResponse = models.ProductResponse{
				ID:   item.ProductID,
				Name: "Unknown Product",
			}
		}

		items[i] = models.OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Product:   productResponse,
		}
	}

	return &models.OrderResponse{
		ID:         order.ID,
		UserID:     order.UserID,
		Status:     order.Status,
		TotalPrice: order.TotalPrice,
		Items:      items,
		CreatedAt:  order.CreatedAt,
	}
}
