# Taskw - Go API Code Generator

A CLI tool for automatically generating Fiber routes and Wire dependency injection code from annotations. Taskw eliminates boilerplate and keeps your API handlers focused on business logic.

## What Problem Does Taskw Solve?

When building Go APIs, you typically write:

1. **Handler functions** with route annotations based on [Swaggo](https://github.com/swaggo/swag) (`@Router /api/users [get]`)
2. **Provider functions** for dependency injection (`func Provide<FUNCTION_NAME>()`)
3. **Boilerplate route registration** (manually wiring routes to handlers)
4. **Boilerplate dependency wiring** (manually connecting all providers)

Taskw automatically generates #3 and #4 by scanning your code for annotations and provider functions.

## Features (v1)

**🎯 Focused Stack:**

- **Web Framework**: Fiber v2 only
- **Dependency Injection**: Google Wire only
- **Annotations**: Swaggo (@Router) only
- **Language**: Go

**🚀 What It Generates:**

- Auto-route registration from `@Router` annotations
- Wire provider sets from `Provide*` functions
- Type-safe dependency injection code
- Development-friendly watching and regeneration

## Installation

```bash
go install github.com/nkaewam/taskw@latest
```

## Quick Start

### 1. Initialize in your project

```bash
cd your-go-api
taskw init
```

This creates a `taskw.yaml` config file:

```yaml
version: "1.0"
project:
  module: "github.com/user/your-go-api"
paths:
  scan_dirs:
    - "./internal/api"
    - "./internal/handlers"
  output_dir: "./internal/api"
generation:
  routes:
    enabled: true
    output_file: "routes_gen.go"
  dependencies:
    enabled: true
    output_file: "dependencies_gen.go"
```

### 2. Write handlers with annotations

```go
// internal/api/user/handler.go
package user

import "github.com/gofiber/fiber/v2"

type Handler struct {
    service *Service
}

// ProvideHandler creates a new user handler
func ProvideHandler(service *Service) *Handler {
    return &Handler{service: service}
}

// GetUsers retrieves all users
// @Summary Get all users
// @Description Get a list of all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} User
// @Router /api/v1/users [get]
func (h *Handler) GetUsers(c *fiber.Ctx) error {
    users, err := h.service.GetAll()
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(users)
}

// CreateUser creates a new user
// @Summary Create user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User data"
// @Success 201 {object} User
// @Router /api/v1/users [post]
func (h *Handler) CreateUser(c *fiber.Ctx) error {
    // Implementation here
    return c.SendStatus(201)
}
```

### 3. Write services with providers

```go
// internal/api/user/service.go
package user

type Service struct {
    repo Repository
}

// ProvideService creates a new user service
func ProvideService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) GetAll() ([]User, error) {
    return s.repo.FindAll()
}
```

### 4. Generate code

```bash
taskw generate
```

This generates:

- `internal/api/routes_gen.go` - Auto-route registration
- `internal/api/dependencies_gen.go` - Wire provider sets

### 5. Use generated code

```go
// cmd/server/main.go
package main

import (
    "github.com/gofiber/fiber/v2"
    "your-module/internal/api"
)

func main() {
    app := fiber.New()

    // Wire generates this function
    server, cleanup, err := api.InitializeServer()
    if err != nil {
        panic(err)
    }
    defer cleanup()

    // Auto-generated route registration
    if err := server.RegisterRoutes(app); err != nil {
        panic(err)
    }

    app.Listen(":8080")
}
```

## Commands

```bash
# Initialize taskw in existing project
taskw init

# Generate all code
taskw generate

# Generate specific components
taskw generate routes      # Only route registration
taskw generate deps        # Only dependency injection

# Watch for changes during development
taskw watch

# Debug what Taskw finds
taskw scan                 # Show discovered handlers and providers
```

## Configuration Reference

```yaml
# taskw.yaml
version: "1.0"

# Project information
project:
  module: "github.com/user/my-api" # Go module from go.mod

# File paths
paths:
  scan_dirs: # Directories to scan
    - "./internal/api"
    - "./internal/handlers"
    - "./pkg/handlers"
  output_dir: "./internal/api" # Where to write generated files

# Code generation settings
generation:
  routes:
    enabled: true
    output_file: "routes_gen.go" # Generated route registration
  dependencies:
    enabled: true
    output_file: "deps_gen.go" # Generated Wire providers
```

## How It Works

### Route Generation

Taskw scans for:

1. **@Router annotations** in comments above handler methods
2. **Provide\* functions** that return handler types

Generated code:

```go
// routes_gen.go (generated)
func (s *Server) RegisterRoutes(app *fiber.App) {
    app.Get("/api/v1/users", s.userHandler.GetUsers)
    app.Post("/api/v1/users", s.userHandler.CreateUser)
    // ... more routes
}
```

### Dependency Generation

Taskw scans for:

1. **Provider functions** starting with `Provide*`
2. **Return types** and parameter dependencies

Generated code:

```go
// deps_gen.go (generated)
var ProviderSet = wire.NewSet(
    user.ProvideHandler,
    user.ProvideService,
    user.ProvideRepository,
    // ... more providers
)
```

## Integration with Build Tools

### Taskfile.yml

```yaml
version: "3"

tasks:
  generate:
    desc: Generate all code
    cmds:
      - taskw generate
      - go generate ./...

  dev:
    desc: Start development server
    deps: [generate]
    cmds:
      - air

  build:
    desc: Build application
    deps: [generate]
    cmds:
      - go build -o bin/server cmd/server/main.go
```

### Makefile

```makefile
.PHONY: generate dev build

generate:
	taskw generate
	go generate ./...

dev: generate
	air

build: generate
	go build -o bin/server cmd/server/main.go
```

### Pre-commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-commit
taskw generate
git add internal/api/*_gen.go
```

## Examples

### Complete Project Structure

```
my-api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── server.go
│   │   ├── wire.go              # Manual Wire config
│   │   ├── routes_gen.go        # Generated by Taskw
│   │   ├── deps_gen.go          # Generated by Taskw
│   │   └── wire_gen.go          # Generated by Wire
│   ├── user/
│   │   ├── handler.go           # Has @Router + ProvideHandler
│   │   ├── service.go           # Has ProvideService
│   │   └── repository.go        # Has ProvideRepository
│   └── product/
│       ├── handler.go
│       ├── service.go
│       └── repository.go
├── taskw.yaml                   # Taskw config
├── Taskfile.yml                 # Build config
└── go.mod
```

### Wire Integration

```go
// internal/api/wire.go
//go:build wireinject

package api

import (
    "github.com/google/wire"
    "github.com/gofiber/fiber/v2"
    "go.uber.org/zap"
)

// ProviderSet combines manual + generated providers
var ProviderSet = wire.NewSet(
    // Manual providers
    provideLogger,
    provideFiberApp,
    // Generated providers (from deps_gen.go)
    GeneratedProviderSet,
    // Server
    NewServer,
)

func InitializeServer() (*Server, func(), error) {
    wire.Build(ProviderSet)
    return &Server{}, nil, nil
}

func provideLogger() *zap.Logger {
    logger, _ := zap.NewProduction()
    return logger
}

func provideFiberApp() *fiber.App {
    return fiber.New()
}
```

## Migration from Existing Projects

If you already have manual route/DI code:

### 1. Install and initialize

```bash
taskw init
```

### 2. Add @Router annotations to existing handlers

```go
// Before
func (h *UserHandler) GetUsers(c *fiber.Ctx) error { ... }

// After
// @Router /api/v1/users [get]
func (h *UserHandler) GetUsers(c *fiber.Ctx) error { ... }
```

### 3. Add Provide\* functions

```go
// Add to existing files
func ProvideUserHandler(service *UserService) *UserHandler {
    return &UserHandler{service: service}
}
```

### 4. Generate and replace manual code

```bash
taskw generate
# Replace manual route registration with generated version
# Replace manual Wire providers with generated version
```

## Development

### Watch Mode

```bash
# Regenerate on file changes
taskw watch

# Combine with Air for live reload
air &
taskw watch &
```

### Debugging

```bash
# See what Taskw discovers
taskw scan

# Output:
# 📁 Scanning ./internal/api
# 🔍 Found handlers:
#   - user.ProvideHandler -> *user.Handler
#   - product.ProvideHandler -> *product.Handler
# 🔍 Found routes:
#   - GetUsers: GET /api/v1/users
#   - CreateUser: POST /api/v1/users
# 📁 Scanning ./internal/handlers
#   (no handlers found)
```

## Troubleshooting

### Common Issues

**Taskw doesn't find my handlers**

- Check `scan_dirs` in `taskw.yaml`
- Ensure files end with `handler.go`
- Ensure functions start with `Provide*`

**Generated routes don't work**

- Check @Router annotation syntax
- Ensure handler methods match `func (h *Handler) Method(c *fiber.Ctx) error`

**Wire compilation fails**

- Check provider function signatures
- Ensure all dependencies have providers
- Run `go generate ./...` after `taskw generate`

**Generated code has wrong imports**

- Check `project.module` in `taskw.yaml` matches `go.mod`

### Debug Output

```bash
# Verbose mode
taskw generate --verbose

# Show generated file contents
taskw generate --dry-run
```

## Roadmap

**v1.1**

- Watch mode (`taskw watch`)
- Template customization
- Better error messages

**v2.0**

- Support for Gin framework
- Support for Fx dependency injection
- Plugin system

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## License

MIT License - see LICENSE file for details.

---

**Taskw** - From manual boilerplate to auto-generated bliss 🚀
