package order

import (
	"fmt"
	"sync"
	"time"

	"github.com/example/ecommerce-api/internal/models"
	"github.com/google/uuid"
)

// Repository handles order data persistence
type Repository struct {
	mu         sync.RWMutex
	orders     map[uuid.UUID]*models.Order
	orderItems map[uuid.UUID]*models.OrderItem
}

// ProvideRepository creates a new order repository
func ProvideRepository() *Repository {
	return &Repository{
		orders:     make(map[uuid.UUID]*models.Order),
		orderItems: make(map[uuid.UUID]*models.OrderItem),
	}
}

// CreateOrder creates a new order with items
func (r *Repository) CreateOrder(order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order.ID = uuid.New()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	// Create order items
	for i := range order.Items {
		order.Items[i].ID = uuid.New()
		order.Items[i].OrderID = order.ID
		r.orderItems[order.Items[i].ID] = &order.Items[i]
	}

	r.orders[order.ID] = order
	return nil
}

// GetOrderByID retrieves an order by ID with its items
func (r *Repository) GetOrderByID(id uuid.UUID) (*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return nil, fmt.Errorf("order with ID %s not found", id)
	}

	// Load order items
	var items []models.OrderItem
	for _, item := range r.orderItems {
		if item.OrderID == id {
			items = append(items, *item)
		}
	}

	// Create a copy with items
	orderWithItems := *order
	orderWithItems.Items = items

	return &orderWithItems, nil
}

// GetOrdersByUserID retrieves all orders for a specific user
func (r *Repository) GetOrdersByUserID(userID uuid.UUID) ([]*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userOrders []*models.Order
	for _, order := range r.orders {
		if order.UserID == userID {
			// Load order items
			var items []models.OrderItem
			for _, item := range r.orderItems {
				if item.OrderID == order.ID {
					items = append(items, *item)
				}
			}

			// Create a copy with items
			orderWithItems := *order
			orderWithItems.Items = items
			userOrders = append(userOrders, &orderWithItems)
		}
	}

	return userOrders, nil
}

// GetAllOrders retrieves all orders
func (r *Repository) GetAllOrders() ([]*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders := make([]*models.Order, 0, len(r.orders))
	for _, order := range r.orders {
		// Load order items
		var items []models.OrderItem
		for _, item := range r.orderItems {
			if item.OrderID == order.ID {
				items = append(items, *item)
			}
		}

		// Create a copy with items
		orderWithItems := *order
		orderWithItems.Items = items
		orders = append(orders, &orderWithItems)
	}

	return orders, nil
}

// UpdateOrderStatus updates the status of an order
func (r *Repository) UpdateOrderStatus(id uuid.UUID, status models.OrderStatus) (*models.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, exists := r.orders[id]
	if !exists {
		return nil, fmt.Errorf("order with ID %s not found", id)
	}

	// Create a copy to avoid modifying the original
	updatedOrder := *order
	updatedOrder.Status = status
	updatedOrder.UpdatedAt = time.Now()

	r.orders[id] = &updatedOrder

	// Load order items for the response
	var items []models.OrderItem
	for _, item := range r.orderItems {
		if item.OrderID == id {
			items = append(items, *item)
		}
	}
	updatedOrder.Items = items

	return &updatedOrder, nil
}

// DeleteOrder deletes an order and its items
func (r *Repository) DeleteOrder(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[id]; !exists {
		return fmt.Errorf("order with ID %s not found", id)
	}

	// Delete order items first
	for itemID, item := range r.orderItems {
		if item.OrderID == id {
			delete(r.orderItems, itemID)
		}
	}

	// Delete the order
	delete(r.orders, id)
	return nil
}

// GetOrdersByStatus retrieves orders by status
func (r *Repository) GetOrdersByStatus(status models.OrderStatus) ([]*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var statusOrders []*models.Order
	for _, order := range r.orders {
		if order.Status == status {
			// Load order items
			var items []models.OrderItem
			for _, item := range r.orderItems {
				if item.OrderID == order.ID {
					items = append(items, *item)
				}
			}

			// Create a copy with items
			orderWithItems := *order
			orderWithItems.Items = items
			statusOrders = append(statusOrders, &orderWithItems)
		}
	}

	return statusOrders, nil
}

// GetOrderItems retrieves all items for a specific order
func (r *Repository) GetOrderItems(orderID uuid.UUID) ([]models.OrderItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []models.OrderItem
	for _, item := range r.orderItems {
		if item.OrderID == orderID {
			items = append(items, *item)
		}
	}

	return items, nil
}
