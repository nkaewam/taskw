package generation

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/nkaewam/taskw/internal/cli/file"
	"github.com/nkaewam/taskw/internal/cli/ui"
	"github.com/nkaewam/taskw/internal/config"
	"github.com/nkaewam/taskw/internal/generator"
	"github.com/nkaewam/taskw/internal/scanner"
)

// Service handles code generation operations
type Service interface {
	// GenerateAll generates routes, dependencies, and swagger documentation
	GenerateAll() error
	// GenerateRoutes generates only route registration code
	GenerateRoutes() error
	// GenerateDependencies generates only dependency injection code
	GenerateDependencies() error
	// GenerateSwagger generates swagger documentation
	GenerateSwagger() error
}

// service implements Service interface
type service struct {
	config      *config.Config
	scanner     *scanner.Scanner
	ui          ui.Service
	fileService file.Service
}

// ProvideGenerationService creates a new generation service
// @Provider
func ProvideGenerationService(config *config.Config, uiService ui.Service, fileService file.Service) Service {
	return &service{
		config:      config,
		scanner:     scanner.NewScanner(config),
		ui:          uiService,
		fileService: fileService,
	}
}

// GenerateAll generates routes, dependencies, and swagger documentation
func (s *service) GenerateAll() error {
	if s.config.Generation.Routes.Enabled {
		if err := s.GenerateRoutes(); err != nil {
			return err
		}
	}
	if s.config.Generation.Dependencies.Enabled {
		if err := s.GenerateDependencies(); err != nil {
			return err
		}
	}

	// Generate Swagger documentation
	return s.GenerateSwagger()
}

// GenerateRoutes generates only route registration code
func (s *service) GenerateRoutes() error {
	if !s.config.Generation.Routes.Enabled {
		return nil
	}

	stopSpinner := s.ui.ShowSpinner("Generating routes...")

	handlers, routes, err := s.scanner.ScanRoutes(s.config.Paths.ScanDirs)
	if err != nil {
		stopSpinner("Error scanning routes")
		return fmt.Errorf("error scanning routes: %w", err)
	}

	if len(handlers) == 0 {
		stopSpinner("No handlers found")
		return nil
	}

	if len(routes) == 0 {
		stopSpinner("No @Router annotations found")
		return nil
	}

	// Generate routes using the RouteGenerator
	routeGen := generator.NewRouteGenerator(s.config)
	if err := routeGen.GenerateRoutes(handlers, routes); err != nil {
		stopSpinner("Error generating routes")
		return fmt.Errorf("error generating routes: %w", err)
	}

	outputPath := filepath.Join(s.config.Paths.OutputDir, s.config.Generation.Routes.OutputFile)
	stopSpinner("Routes generated successfully")
	fmt.Printf("  • Found %d handlers and %d routes\n", len(handlers), len(routes))
	fmt.Printf("  • Generated: %s\n", outputPath)

	return nil
}

// GenerateDependencies generates only dependency injection code
func (s *service) GenerateDependencies() error {
	if !s.config.Generation.Dependencies.Enabled {
		return nil
	}

	stopSpinner := s.ui.ShowSpinner("Generating dependencies...")

	providers, err := s.scanner.ScanProviders(s.config.Paths.ScanDirs)
	if err != nil {
		stopSpinner("Error scanning providers")
		return fmt.Errorf("error scanning providers: %w", err)
	}

	if len(providers) == 0 {
		stopSpinner("No provider functions found")
		return nil
	}

	// Generate dependencies using the DependencyGenerator
	depGen := generator.NewDependencyGenerator(s.config)
	if err := depGen.GenerateDependencies(providers); err != nil {
		stopSpinner("Error generating dependencies")
		return fmt.Errorf("error generating dependencies: %w", err)
	}

	outputPath := filepath.Join(s.config.Paths.OutputDir, s.config.Generation.Dependencies.OutputFile)
	stopSpinner("Dependencies generated successfully")
	fmt.Printf("  • Found %d providers\n", len(providers))
	fmt.Printf("  • Generated: %s\n", outputPath)

	return nil
}

// GenerateSwagger generates swagger documentation
func (s *service) GenerateSwagger() error {
	stopSpinner := s.ui.ShowSpinner("Generating Swagger documentation...")

	// Check if swag command is available
	if !s.fileService.IsCommandAvailable("swag") {
		stopSpinner("Installing swag command...")
		installSpinner := s.ui.ShowSpinner("Installing swag...")

		if err := s.fileService.InstallSwag(); err != nil {
			installSpinner("Failed to install swag")
			fmt.Printf("  Please install manually: go install github.com/swaggo/swag/cmd/swag@latest\n")
			return nil
		}
		installSpinner("swag installed successfully")
	}

	// Generate swagger docs
	mainFile := s.fileService.FindMainFile()
	if mainFile == "" {
		stopSpinner("Could not find main.go file for swagger generation")
		return nil
	}

	docsDir := "docs"
	cmd := exec.Command("swag", "init", "-g", mainFile, "-o", docsDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		stopSpinner("Error generating swagger docs")
		fmt.Printf("Output: %s\n", string(output))
		return fmt.Errorf("error generating swagger docs: %w", err)
	}

	stopSpinner(fmt.Sprintf("Swagger documentation generated successfully at %s/", docsDir))
	return nil
}
