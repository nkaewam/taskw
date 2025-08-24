package api

import (
	"github.com/example/ecommerce-api/internal/health"
	"github.com/example/ecommerce-api/internal/order"
	"github.com/example/ecommerce-api/internal/product"
	"github.com/example/ecommerce-api/internal/user"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Server holds all the handlers and dependencies
type Server struct {
	logger         *zap.Logger
	healthHandler  *health.Handler
	userHandler    *user.Handler
	productHandler *product.Handler
	orderHandler   *order.Handler
}

// ProvideServer creates a new server with all dependencies
func ProvideServer(
	logger *zap.Logger,
	healthHandler *health.Handler,
	userHandler *user.Handler,
	productHandler *product.Handler,
	orderHandler *order.Handler,
) *Server {
	return &Server{
		logger:         logger,
		healthHandler:  healthHandler,
		userHandler:    userHandler,
		productHandler: productHandler,
		orderHandler:   orderHandler,
	}
}

// ProvideFiberApp creates a new Fiber application
func ProvideFiberApp() *fiber.App {
	return fiber.New(fiber.Config{
		AppName: "E-commerce API",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})
}
