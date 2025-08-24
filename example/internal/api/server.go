package api

import (
	"github.com/example/ecommerce-api/internal/order"
	"github.com/example/ecommerce-api/internal/product"
	"github.com/example/ecommerce-api/internal/user"
	"go.uber.org/zap"
)

// Server holds all the handlers and dependencies
type Server struct {
	logger         *zap.Logger
	userHandler    *user.Handler
	productHandler *product.Handler
	orderHandler   *order.Handler
}

// ProvideServer creates a new server with all dependencies
func ProvideServer(
	logger *zap.Logger,
	userHandler *user.Handler,
	productHandler *product.Handler,
	orderHandler *order.Handler,
) *Server {
	return &Server{
		logger:         logger,
		userHandler:    userHandler,
		productHandler: productHandler,
		orderHandler:   orderHandler,
	}
}
