package order

import (
	"github.com/example/ecommerce-api/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for order operations
type Handler struct {
	service *Service
}

// ProvideHandler creates a new order handler
func ProvideHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetOrders retrieves orders with optional filters
// @Summary Get orders
// @Description Get orders with optional filtering by user or status
// @Tags orders
// @Accept json
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param status query string false "Filter by status (pending, confirmed, shipped, delivered, cancelled)"
// @Success 200 {array} models.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders [get]
func (h *Handler) GetOrders(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	statusStr := c.Query("status")

	// Filter by user ID
	if userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid user ID format",
			})
		}

		orders, err := h.service.GetOrdersByUser(userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(orders)
	}

	// Filter by status
	if statusStr != "" {
		status := models.OrderStatus(statusStr)
		if !isValidOrderStatus(status) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid status. Valid values: pending, confirmed, shipped, delivered, cancelled",
			})
		}

		orders, err := h.service.GetOrdersByStatus(status)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(orders)
	}

	// Get all orders
	orders, err := h.service.GetAllOrders()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(orders)
}

// GetOrder retrieves an order by ID
// @Summary Get order by ID
// @Description Get a specific order by its ID
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} models.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/{id} [get]
func (h *Handler) GetOrder(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order ID is required",
		})
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid order ID format",
		})
	}

	order, err := h.service.GetOrder(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(order)
}

// CreateOrder creates a new order
// @Summary Create a new order
// @Description Create a new order for a user
// @Tags orders
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Param order body models.CreateOrderRequest true "Order creation data"
// @Success 201 {object} models.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders [post]
func (h *Handler) CreateOrder(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id query parameter is required",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user ID format",
		})
	}

	var req models.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	order, err := h.service.CreateOrder(userID, &req)
	if err != nil {
		if contains(err.Error(), "insufficient stock") || contains(err.Error(), "invalid") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

// UpdateOrderStatus updates the status of an order
// @Summary Update order status
// @Description Update the status of an existing order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param status body models.UpdateOrderStatusRequest true "Status update data"
// @Success 200 {object} models.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/{id}/status [put]
func (h *Handler) UpdateOrderStatus(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order ID is required",
		})
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid order ID format",
		})
	}

	var req models.UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if !isValidOrderStatus(req.Status) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid status. Valid values: pending, confirmed, shipped, delivered, cancelled",
		})
	}

	order, err := h.service.UpdateOrderStatus(id, &req)
	if err != nil {
		if contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if contains(err.Error(), "invalid") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(order)
}

// CancelOrder cancels an order
// @Summary Cancel order
// @Description Cancel an existing order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} models.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/{id}/cancel [post]
func (h *Handler) CancelOrder(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order ID is required",
		})
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid order ID format",
		})
	}

	if err := h.service.CancelOrder(id); err != nil {
		if contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		if contains(err.Error(), "invalid") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "order cancelled successfully",
	})
}

// GetUserOrders retrieves all orders for a specific user
// @Summary Get user orders
// @Description Get all orders for a specific user
// @Tags orders
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {array} models.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{user_id}/orders [get]
func (h *Handler) GetUserOrders(c *fiber.Ctx) error {
	userIDStr := c.Params("user_id")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user ID is required",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user ID format",
		})
	}

	orders, err := h.service.GetOrdersByUser(userID)
	if err != nil {
		if contains(err.Error(), "invalid user") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(orders)
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[len(s)-len(substr):] == substr || 
		   (len(s) > len(substr) && s[:len(substr)] == substr) ||
		   (len(s) > len(substr)*2 && s[len(s)/2-len(substr)/2:len(s)/2+len(substr)/2] == substr))
}

func isValidOrderStatus(status models.OrderStatus) bool {
	validStatuses := []models.OrderStatus{
		models.OrderStatusPending,
		models.OrderStatusConfirmed,
		models.OrderStatusShipped,
		models.OrderStatusDelivered,
		models.OrderStatusCancelled,
	}

	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}

	return false
}
