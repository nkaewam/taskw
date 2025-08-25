package project

import (
	"fmt"

	"github.com/nkaewam/taskw/internal/cli/ui"
	"github.com/nkaewam/taskw/internal/generator"
)

// Service handles project initialization and scaffolding
type Service interface {
	// InitProject creates a new project with full scaffolding
	InitProject(projectPath, module, projectName string) error
	// ValidateModule validates that the module path is a proper Go module format
	ValidateModule(module string) error
	// ExtractProjectName extracts the project name from a module path
	ExtractProjectName(module string) string
	// ValidateProjectPath validates that a project directory can be created
	ValidateProjectPath(projectPath string) error
}

// service implements Service interface
type service struct {
	ui ui.Service
}

// ProvideProjectService creates a new project service
// @Provider
func ProvideProjectService(uiService ui.Service) Service {
	return &service{
		ui: uiService,
	}
}

// InitProject creates a new project with full scaffolding
func (s *service) InitProject(projectPath, module, projectName string) error {
	// Validate project directory
	initGen := generator.NewInitGenerator()
	if err := initGen.ValidateProjectPath(projectPath); err != nil {
		return fmt.Errorf("invalid project path: %w", err)
	}

	// Generate the project
	if err := initGen.InitProject(projectPath, module, projectName); err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}

	// Success message
	fmt.Println("\n🎉 Project scaffolded successfully!")
	fmt.Printf("📁 Created in: %s/\n", projectPath)
	fmt.Printf("📦 Module: %s\n", module)

	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  go mod tidy")
	fmt.Println("  task setup           # Install dependencies and generate code")
	fmt.Println("  task dev             # Start development server with live reload")
	fmt.Println("\nOr run manually:")
	fmt.Println("  taskw generate       # Generate routes and dependencies")
	fmt.Println("  go run cmd/server/main.go  # Start the server")

	return nil
}

// ValidateModule validates that the module path is a proper Go module format
func (s *service) ValidateModule(module string) error {
	return ui.ValidateModule(module)
}

// ExtractProjectName extracts the project name from a module path
func (s *service) ExtractProjectName(module string) string {
	return ui.ExtractProjectName(module)
}

// ValidateProjectPath validates that a project directory can be created
func (s *service) ValidateProjectPath(projectPath string) error {
	initGen := generator.NewInitGenerator()
	return initGen.ValidateProjectPath(projectPath)
}
