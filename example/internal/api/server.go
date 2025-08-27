package api

import (
	"github.com/gofiber/fiber/v2"
)

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
