# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Taskw is a Go CLI tool that generates code for Go APIs using the Fiber web framework, Wire dependency injection, and Swaggo annotations. It eliminates boilerplate by scanning your code for annotations and provider functions to automatically generate route registration and dependency injection code.

**Core Technologies:**
- Go 1.24.1
- Fiber v2 (web framework)
- Google Wire (dependency injection)
- Swaggo (OpenAPI/Swagger annotations)
- Cobra (CLI framework)
- Viper (configuration management)

## Common Development Commands

### Build and Install
```bash
# Build the binary
task build

# Install development version to PATH
task install-dev

# Install production version to PATH  
task install

# Build manually
go build -o bin/taskw main.go
```

### Testing
```bash
# Run all tests
task test

# Run end-to-end tests
task test-e2e

# Run tests manually
go test -v ./...
go test -v ./test/e2e/...
```

### Code Generation
The project uses Wire for dependency injection. After making changes to providers:
```bash
# Generate Wire code
go generate ./...

# Or specifically for CLI services
go generate ./internal/cli/
```

### Using Taskw CLI
```bash
# Initialize new project
taskw init [module-name]

# Generate all code  
taskw generate
taskw generate all

# Generate specific components
taskw generate routes
taskw generate deps

# Scan and preview what will be generated
taskw scan

# Clean generated files
taskw clean
```

## Architecture Overview

### Project Structure
```
taskw/
├── cmd/taskw/           # CLI entry point (Cobra commands)
├── internal/
│   ├── cli/             # CLI service layer with Wire DI
│   ├── config/          # Configuration management (Viper)
│   ├── scanner/         # Code scanning (AST parsing + file filtering)  
│   └── generator/       # Code generation (templates + formatting)
├── main.go              # Application entry point
└── taskw.yaml          # Configuration file
```

### Dependency Injection Architecture

The CLI uses Wire for dependency injection with the following pattern:

1. **Services**: Located in `internal/cli/*/service.go` files
2. **Wire Configuration**: `internal/cli/wire.go` defines providers
3. **Generated Code**: `internal/cli/wire_gen.go` and `dependencies_gen.go`
4. **Container Pattern**: `cli.Container` holds all injected services

**Key Services:**
- `ui.Service`: User interface and spinner management  
- `project.Service`: Project scaffolding and validation
- `scan.Service`: Code scanning and analysis
- `generation.Service`: Code generation orchestration
- `clean.Service`: Generated file cleanup
- `file.Service`: File system operations

### Route Generation Architecture

The route generator handles complex routing patterns with specificity-based ordering:

**Route Processing Pipeline:**
1. **Path Conversion**: Converts OpenAPI `{param}` format to Fiber `:param` format
2. **Specificity Scoring**: Routes are scored and sorted (more specific routes first)
3. **Package Organization**: Routes are grouped by package but sorted globally
4. **Import Path Resolution**: Automatically derives handler import paths from package structure

**Key Components:**
- `RouteGenerator`: Main route generation orchestrator
- `HandlerInfo`: Represents handler dependency injection information with full package paths
- **Specificity Algorithm**: Uses segment count and parameter penalties for routing order
- **Template System**: Embedded templates with Go formatting and import management

**Route Specificity Rules:**
- Longer paths get higher scores (1000 points per segment)  
- Static segments add 100 points each
- Parameters (`:param`) subtract 100 points each
- Routes with higher scores are registered first to prevent conflicts

### Scanning Architecture

The scanner uses a hybrid approach with parallel processing:
- **File Filtering**: Identifies candidate files quickly
- **AST Parsing**: Parallel processing with semaphore-controlled goroutines (max 10)
- **Error Collection**: Non-fatal errors are collected in ScanResult.Errors

**Key Scanner Types:**
- `scanner.HandlerFunction`: Represents provider functions
- `scanner.RouteMapping`: Represents route annotations with package paths
- `scanner.ProviderFunction`: Represents Wire provider functions
- `scanner.ScanResult`: Combined scan results with error tracking
- `scanner.ScanError`: Error information with file path and type
- `scanner.ScanStatistics`: Performance metrics for debugging

**Scanner Components:**
- `Scanner`: Main hybrid scanner orchestrating file filtering and AST parsing
- `ASTScanner`: Handles Go AST parsing for annotations and providers
- `FileFilter`: Optimizes file discovery before AST parsing

## Configuration Management

Taskw uses Viper for configuration with the following hierarchy:
1. Command line flags (`--config`)
2. `taskw.yaml` file
3. Default values

**Key Configuration Sections:**
- `paths.scan_dirs`: Directories to scan for code
- `paths.output_dir`: Where to generate files  
- `generation.routes.enabled`: Enable route generation
- `generation.dependencies.enabled`: Enable dependency generation

## Code Style and Patterns

### Service Pattern
All CLI functionality is organized into services with this pattern:
```go
type Service interface {
    // Interface defines contract
}

type serviceImpl struct {
    // Dependencies injected via constructor
}

func ProvideService(deps...) Service {
    return &serviceImpl{...}
}
```

### Error Handling
Use wrapped errors with context:
```go
return fmt.Errorf("failed to scan directory %s: %w", dir, err)
```

### Template Usage
Code generation uses embedded templates in `internal/generator/templates/`:
- Templates are embedded using `//go:embed`
- Use `text/template` for Go code generation
- Apply `go/format` to generated code

### Wire Integration
When adding new services:
1. Create service interface and implementation
2. Add `Provide*` function
3. Add to `internal/cli/dependencies_gen.go` (or let Taskw generate it)
4. Run `go generate ./internal/cli/`

## Testing Approach

- **Unit Tests**: Standard Go tests in `*_test.go` files
- **E2E Tests**: Located in `test/e2e/` directory  
- **Integration Tests**: Test complete CLI workflows

## Important Implementation Details

### File Generation Safety
- Generated files include header comments marking them as generated
- Clean command only removes files with generation markers
- Templates handle Go imports and formatting automatically

### CLI UX Patterns  
- Use spinners for long-running operations
- Provide clear error messages with actionable suggestions
- Support both interactive and non-interactive modes

### Performance Considerations
- **Parallel Scanning**: Uses semaphore-controlled goroutines (max 10) for file processing
- **File Filtering**: Pre-filters candidates before expensive AST parsing
- **Route Specificity**: Complex scoring algorithm ensures correct route registration order
- **Error Resilience**: Non-fatal scan errors don't stop processing of other files
- **Generated Code Caching**: Files are only regenerated when source changes
- **Wire Compile-Time Safety**: Dependency injection validated at compile time

### Route Generation Specifics
- **Path Parameter Conversion**: Automatically converts `{id}` to `:id` for Fiber compatibility
- **Import Path Derivation**: Uses project module config to build correct import paths
- **Handler Reference Resolution**: Converts handler refs like `userHandler.GetUsers` to `ar.userHandler.GetUsers`
- **Deterministic Ordering**: Routes and handlers are sorted for consistent generated code

## Common Tasks for Contributors

### Adding a New CLI Command
1. Add command definition to `cmd/taskw/main.go`
2. Create service interface if complex logic needed
3. Add service implementation in `internal/cli/[service]/`
4. Wire dependencies in `internal/cli/wire.go`
5. Generate Wire code with `go generate ./internal/cli/`

### Modifying Code Generation
1. Update scanner types in `internal/scanner/types.go`  
2. Modify templates in `internal/generator/templates/`
3. Update generator logic in `internal/generator/`
4. Consider route specificity impact if changing route handling
5. Test with `taskw scan` and `taskw generate`

### Working with Scanner Changes
1. Understand parallel processing model - changes affect goroutine safety
2. Update `ScanResult` struct if adding new scan data types
3. Handle errors gracefully using `ScanError` collection pattern
4. Test performance impact with `GetStatistics()` for large codebases
5. Consider AST vs file filtering trade-offs for new features

### Adding Configuration Options
1. Update config structs in `internal/config/config.go`
2. Add Viper defaults in `setDefaults()`
3. Update example `taskw.yaml` in documentation