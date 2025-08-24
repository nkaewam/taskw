package product

import (
	"strconv"

	"github.com/example/ecommerce-api/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for product operations
type Handler struct {
	service *Service
}

// ProvideHandler creates a new product handler
func ProvideHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetProducts retrieves all products
// @Summary Get all products
// @Description Get a list of all products in the system
// @Tags products
// @Accept json
// @Produce json
// @Param category query string false "Filter by category ID"
// @Success 200 {array} models.ProductResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/products [get]
func (h *Handler) GetProducts(c *fiber.Ctx) error {
	categoryID := c.Query("category")
	
	if categoryID != "" {
		// Filter by category
		catID, err := uuid.Parse(categoryID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid category ID format",
			})
		}

		products, err := h.service.GetProductsByCategory(catID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(products)
	}

	products, err := h.service.GetProducts()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(products)
}

// GetProduct retrieves a product by ID
// @Summary Get product by ID
// @Description Get a specific product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} models.ProductResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/products/{id} [get]
func (h *Handler) GetProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product ID is required",
		})
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product ID format",
		})
	}

	product, err := h.service.GetProduct(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(product)
}

// CreateProduct creates a new product
// @Summary Create a new product
// @Description Create a new product in the system
// @Tags products
// @Accept json
// @Produce json
// @Param product body models.CreateProductRequest true "Product creation data"
// @Success 201 {object} models.ProductResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/products [post]
func (h *Handler) CreateProduct(c *fiber.Ctx) error {
	var req models.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	product, err := h.service.CreateProduct(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}

// UpdateProduct updates an existing product
// @Summary Update product
// @Description Update an existing product's information
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body models.UpdateProductRequest true "Product update data"
// @Success 200 {object} models.ProductResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/products/{id} [put]
func (h *Handler) UpdateProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product ID is required",
		})
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product ID format",
		})
	}

	var req models.UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	product, err := h.service.UpdateProduct(id, &req)
	if err != nil {
		if contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(product)
}

// DeleteProduct deletes a product
// @Summary Delete product
// @Description Delete a product from the system
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/products/{id} [delete]
func (h *Handler) DeleteProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product ID is required",
		})
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product ID format",
		})
	}

	if err := h.service.DeleteProduct(id); err != nil {
		if contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetCategories retrieves all product categories
// @Summary Get all categories
// @Description Get a list of all product categories
// @Tags products
// @Accept json
// @Produce json
// @Success 200 {array} models.Category
// @Failure 500 {object} map[string]string
// @Router /api/v1/categories [get]
func (h *Handler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.service.GetCategories()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(categories)
}

// CheckStock checks product stock availability
// @Summary Check product stock
// @Description Check if enough stock is available for a product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param quantity query int true "Quantity to check"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/products/{id}/stock [get]
func (h *Handler) CheckStock(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product ID is required",
		})
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product ID format",
		})
	}

	quantityStr := c.Query("quantity")
	if quantityStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "quantity parameter is required",
		})
	}

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid quantity format",
		})
	}

	if quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "quantity must be greater than 0",
		})
	}

	available, err := h.service.CheckStock(id, quantity)
	if err != nil {
		if contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"available": available,
		"requested": quantity,
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr || 
		   len(s) > len(substr) && s[:len(substr)] == substr ||
		   len(s) > len(substr)*2 && s[len(s)/2-len(substr)/2:len(s)/2+len(substr)/2] == substr
}
