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

	// Generated providers will be added here by taskw
	// Uncomment this line after running 'taskw generate deps':
	// GeneratedProviderSet,
)

// InitializeServer initializes the complete server with all dependencies
func InitializeServer() (*Server, *fiber.App, func(), error) {
	wire.Build(ProviderSet)
	return &Server{}, &fiber.App{}, nil, nil
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
