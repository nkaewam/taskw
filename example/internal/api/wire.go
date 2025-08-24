//go:build wireinject

package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// ProviderSet will be augmented by taskw generated dependencies
// This only contains infrastructure providers - taskw will add the rest
var ProviderSet = wire.NewSet(
	// Infrastructure providers (manual)
	provideLogger,
	provideFiberApp,

	// Adapters (manual)
	ProvideProductServiceAdapter,
	ProvideUserServiceAdapter,

	// Server (manual)
	ProvideServer,

	// Generated providers added by taskw
	GeneratedProviderSet,
)

// InitializeServer initializes the complete server with all dependencies
func InitializeServer() (*Server, error) {
	wire.Build(ProviderSet)
	return &Server{}, nil
}

// InitializeFiberApp initializes the Fiber app
func InitializeFiberApp() *fiber.App {
	wire.Build(provideFiberApp)
	return &fiber.App{}
}

// provideLogger creates a new zap logger
func provideLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return logger, nil
}

// provideFiberApp creates a new Fiber application
func provideFiberApp() *fiber.App {
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
