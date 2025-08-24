package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nkaewam/taskw/internal/config"
	"github.com/nkaewam/taskw/internal/generator"
	"github.com/nkaewam/taskw/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	configPath string
)

var rootCmd = &cobra.Command{
	Use:   "taskw",
	Short: "Go API Code Generator (Fiber + Wire + Swaggo)",
	Long: `Taskw is a code generator for Go APIs that automatically generates:
- Fiber route registration from handler functions
- Wire dependency injection setup from provider functions  
- Swagger documentation

It scans your Go code for special annotations and generates boilerplate code to wire everything together.`,
}

// Spinner handles animated loading indicators
type Spinner struct {
	chars []string
	delay time.Duration
	done  chan bool
	mu    sync.Mutex
}

func NewSpinner() *Spinner {
	return &Spinner{
		chars: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		delay: 100 * time.Millisecond,
		done:  make(chan bool),
	}
}

func (s *Spinner) Start(message string) {
	go func() {
		i := 0
		for {
			select {
			case <-s.done:
				return
			default:
				s.mu.Lock()
				fmt.Printf("\r%s %s", s.chars[i%len(s.chars)], message)
				s.mu.Unlock()
				i++
				time.Sleep(s.delay)
			}
		}
	}()
}

func (s *Spinner) Stop(message string) {
	s.done <- true
	s.mu.Lock()
	fmt.Printf("\r✔ %s\n", message)
	s.mu.Unlock()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to taskw.yaml config file")

	// Add flags to init command
	initCmd.Flags().Bool("config-only", false, "Only create configuration files (taskw.yaml and .taskwignore)")

	// Setup generate subcommands
	generateCmd.AddCommand(generateAllCmd)
	generateCmd.AddCommand(generateRoutesCmd)
	generateCmd.AddCommand(generateDepsCmd)
	// Set "all" as the default command when just "generate" is called
	generateCmd.Run = generateAllCmd.Run

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(cleanCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var initCmd = &cobra.Command{
	Use:   "init [module]",
	Short: "Initialize a new Taskw project with full scaffold",
	Long: `Initialize a new Taskw project by scaffolding:
- cmd/server/main.go - Main server entry point with Swagger docs
- internal/api/server.go - Server struct and providers
- internal/api/wire.go - Wire dependency injection setup
- internal/health/handler.go - Example health check handler
- .air.toml - Live reload configuration
- Taskfile.yml - Task runner configuration
- taskw.yaml - Taskw configuration
- go.mod - Go module file

Requires a full Go module path (e.g., github.com/user/project-name).

Examples:
  taskw init                                    # Interactive prompt for module
  taskw init github.com/user/my-api             # Create project with specified module
  taskw init --config-only                      # Only create taskw.yaml and .taskwignore`,
	Run: func(cmd *cobra.Command, args []string) {
		handleNewInit(cmd, args)
	},
}

func handleNewInit(cmd *cobra.Command, args []string) {
	configOnly, _ := cmd.Flags().GetBool("config-only")

	if configOnly {
		handleConfigInit(configPath)
		return
	}

	// Full project scaffolding
	var module string
	if len(args) == 0 {
		// Interactive prompt for module
		var err error
		module, err = promptForModule()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		module = args[0]
		// Validate module format
		if err := validateModule(module); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Extract project name from module path
	projectName := extractProjectName(module)
	projectPath := filepath.Join(".", projectName)

	// Validate project directory
	initGen := generator.NewInitGenerator()
	if err := initGen.ValidateProjectPath(projectPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	spinner := NewSpinner()
	spinner.Start(fmt.Sprintf("Creating project %s...", projectName))

	// Generate the project
	if err := initGen.InitProject(projectPath, module, projectName); err != nil {
		spinner.Stop("Project creation failed")
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	spinner.Stop(fmt.Sprintf("Project %s created successfully", projectName))

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
}

func handleConfigInit(configPath string) {
	if configPath == "" {
		configPath = "taskw.yaml"
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config file %s already exists\n", configPath)
	} else {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Error creating config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("● Created taskw.yaml")
		fmt.Printf("  Configured to scan: %v\n", cfg.Paths.ScanDirs)
		fmt.Printf("  Output directory: %s\n", cfg.Paths.OutputDir)
	}

	// Create .taskwignore if it doesn't exist
	if _, err := os.Stat(".taskwignore"); os.IsNotExist(err) {
		filter := scanner.NewFileFilter()
		if err := filter.CreateDefaultTaskwIgnore(); err != nil {
			fmt.Printf("Warning: Could not create .taskwignore: %v\n", err)
		} else {
			fmt.Println("● Created .taskwignore")
		}
	} else {
		fmt.Println("• Using existing .taskwignore")
	}

	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit taskw.yaml to configure scan directories and output settings")
	fmt.Println("  2. Edit .taskwignore to exclude files/directories from scanning")
	fmt.Println("  3. Run 'taskw scan' to preview what will be generated")
	fmt.Println("  4. Run 'taskw generate' to generate code")
}

// promptForModule interactively prompts for a Go module path
func promptForModule() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("🚀 Let's create a new Taskw project!")
	fmt.Println()

	for {
		fmt.Print("Enter Go module path (e.g., github.com/username/my-awesome-api): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		module := strings.TrimSpace(input)
		if module == "" {
			fmt.Println("❌ Module path cannot be empty. Please try again.")
			continue
		}

		if err := validateModule(module); err != nil {
			fmt.Printf("❌ %v Please try again.\n", err)
			continue
		}

		projectName := extractProjectName(module)
		fmt.Printf("✅ Great! Creating project '%s' with module '%s'\n", projectName, module)
		return module, nil
	}
}

// validateModule validates that the module path is a proper Go module format
func validateModule(module string) error {
	// Basic module format validation
	if !strings.Contains(module, "/") {
		return fmt.Errorf("module must contain at least one '/' (e.g., github.com/user/project)")
	}

	// Check for valid module path format
	modulePattern := regexp.MustCompile(`^[a-zA-Z0-9.-]+(\.[a-zA-Z]{2,})?(/[a-zA-Z0-9._-]+)+$`)
	if !modulePattern.MatchString(module) {
		return fmt.Errorf("invalid module format. Use format like: github.com/user/project-name")
	}

	// Extract and validate project name (last part of module)
	projectName := extractProjectName(module)
	return validateProjectName(projectName)
}

// extractProjectName extracts the project name from a module path
func extractProjectName(module string) string {
	parts := strings.Split(module, "/")
	return parts[len(parts)-1]
}

// validateProjectName validates that the project name follows slug-case format
func validateProjectName(name string) error {
	// Check for slug-case format: lowercase letters, numbers, and hyphens only
	// Cannot start or end with hyphen, cannot have consecutive hyphens
	slugPattern := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

	if !slugPattern.MatchString(name) {
		return fmt.Errorf("project name (last part of module) must be in slug-case (lowercase letters, numbers, and hyphens only, e.g., 'my-api')")
	}

	// Additional validation rules
	if len(name) < 2 {
		return fmt.Errorf("project name must be at least 2 characters long")
	}

	if len(name) > 50 {
		return fmt.Errorf("project name must be no longer than 50 characters")
	}

	// Check for reserved names
	reservedNames := []string{
		"api", "app", "main", "src", "lib", "bin", "cmd", "internal",
		"pkg", "test", "tests", "doc", "docs", "build", "dist",
	}

	for _, reserved := range reservedNames {
		if name == reserved {
			return fmt.Errorf("'%s' is a reserved name, please choose a different project name", name)
		}
	}

	return nil
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate code",
	Long: `Generate various types of code from your annotated Go files:
- all: Generate routes and dependencies (default)
- routes: Generate Fiber route registration
- deps/dependencies: Generate Wire dependency injection`,
}

var generateAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Generate routes and dependencies",
	Long:  `Generate both route registration and dependency injection code, plus Swagger documentation.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
		s := scanner.NewScanner(cfg)
		generateAll(s, cfg)
	},
}

var generateRoutesCmd = &cobra.Command{
	Use:   "routes",
	Short: "Generate Fiber route registration",
	Long:  `Generate route registration code from handler functions with @Router annotations.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
		s := scanner.NewScanner(cfg)
		generateRoutes(s, cfg)
	},
}

var generateDepsCmd = &cobra.Command{
	Use:     "deps",
	Aliases: []string{"dependencies"},
	Short:   "Generate Wire dependency injection",
	Long:    `Generate Wire dependency injection setup from provider functions.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
		s := scanner.NewScanner(cfg)
		generateDeps(s, cfg)
	},
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Show what will be generated",
	Long: `Scan the codebase and display what handlers, routes, and providers would be generated.
This is useful for previewing changes before running generate.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleScan(configPath)
	},
}

func handleScan(configPath string) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create scanner
	s := scanner.NewScanner(cfg)

	spinner := NewSpinner()
	spinner.Start("Scanning codebase...")
	fmt.Println("• Using ignore patterns from .taskwignore")

	// Scan all configured directories
	result, err := s.ScanAll()
	if err != nil {
		spinner.Stop("Scan failed")
		fmt.Printf("Error scanning: %v\n", err)
		os.Exit(1)
	}

	spinner.Stop("Codebase scanned successfully")

	// Display results
	stats := s.GetStatistics(result)
	fmt.Printf("\nScan Results:\n")
	fmt.Printf("  • Handlers found: %d\n", stats.HandlersFound)
	fmt.Printf("  • Routes found: %d\n", stats.RoutesFound)
	fmt.Printf("  • Providers found: %d\n", stats.ProvidersFound)
	fmt.Printf("  • Packages scanned: %d\n", stats.PackagesScanned)

	if stats.ErrorsFound > 0 {
		fmt.Printf("  • Errors: %d\n", stats.ErrorsFound)
	}

	// Show detailed results if requested
	if len(result.Handlers) > 0 {
		fmt.Println("\nHandlers:")
		for _, h := range result.Handlers {
			fmt.Printf("  - %s.%s (%s)\n", h.Package, h.FunctionName, h.HandlerName)
		}
	}

	if len(result.Routes) > 0 {
		fmt.Println("\nRoutes:")
		for _, r := range result.Routes {
			fmt.Printf("  - %s %s -> %s\n", r.HTTPMethod, r.Path, r.HandlerRef)
		}
	}

	if len(result.Providers) > 0 {
		fmt.Println("\nProviders:")
		for _, p := range result.Providers {
			fmt.Printf("  - %s() -> %s\n", p.FunctionName, p.ReturnType)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range result.Errors {
			fmt.Printf("  - %s: %s\n", e.FilePath, e.Message)
		}
	}

	// Validate results
	validator := scanner.NewValidator()
	validation := validator.ValidateScanResult(result)

	if validation.HasErrors() {
		fmt.Println("\nValidation Errors:")
		for _, err := range validation.Errors {
			fmt.Printf("  • %s: %s\n", err.Type, err.Message)
		}
	}

	if validation.HasWarnings() {
		fmt.Println("\nValidation Warnings:")
		for _, warn := range validation.Warnings {
			fmt.Printf("  • %s: %s\n", warn.Type, warn.Message)
		}
	}
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove all generated files",
	Long: `Remove all files that were generated by Taskw:
- Route registration files
- Dependency injection files  
- Swagger documentation files

This helps clean up the workspace when regenerating code or switching configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleClean(configPath)
	},
}

func handleClean(configPath string) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	spinner := NewSpinner()
	spinner.Start("Cleaning generated files...")

	var deletedFiles []string
	var skippedFiles []string

	// Clean routes file
	if cfg.Generation.Routes.Enabled {
		routesPath := filepath.Join(cfg.Paths.OutputDir, cfg.Generation.Routes.OutputFile)
		if deleted, err := deleteIfExists(routesPath); err != nil {
			fmt.Printf("• Error deleting %s: %v\n", routesPath, err)
		} else if deleted {
			deletedFiles = append(deletedFiles, routesPath)
		} else {
			skippedFiles = append(skippedFiles, routesPath)
		}
	}

	// Clean dependencies file
	if cfg.Generation.Dependencies.Enabled {
		depsPath := filepath.Join(cfg.Paths.OutputDir, cfg.Generation.Dependencies.OutputFile)
		if deleted, err := deleteIfExists(depsPath); err != nil {
			fmt.Printf("• Error deleting %s: %v\n", depsPath, err)
		} else if deleted {
			deletedFiles = append(deletedFiles, depsPath)
		} else {
			skippedFiles = append(skippedFiles, depsPath)
		}
	}

	// Clean swagger documentation
	docsDir := "docs"
	swaggerFiles := []string{
		filepath.Join(docsDir, "docs.go"),
		filepath.Join(docsDir, "swagger.json"),
		filepath.Join(docsDir, "swagger.yaml"),
	}

	for _, swaggerFile := range swaggerFiles {
		if deleted, err := deleteIfExists(swaggerFile); err != nil {
			fmt.Printf("• Error deleting %s: %v\n", swaggerFile, err)
		} else if deleted {
			deletedFiles = append(deletedFiles, swaggerFile)
		} else {
			skippedFiles = append(skippedFiles, swaggerFile)
		}
	}

	// Try to remove docs directory if it's empty
	if _, err := os.Stat(docsDir); err == nil {
		if err := os.Remove(docsDir); err == nil {
			deletedFiles = append(deletedFiles, docsDir+"/")
		}
		// Ignore error if directory is not empty - that's fine
	}

	spinner.Stop("Clean completed successfully")

	// Report results
	if len(deletedFiles) > 0 {
		fmt.Printf("● Deleted %d files:\n", len(deletedFiles))
		for _, file := range deletedFiles {
			fmt.Printf("  - %s\n", file)
		}
	}

	if len(skippedFiles) > 0 {
		fmt.Printf("• Skipped %d files (not found):\n", len(skippedFiles))
		for _, file := range skippedFiles {
			fmt.Printf("  - %s\n", file)
		}
	}

	if len(deletedFiles) == 0 && len(skippedFiles) == 0 {
		fmt.Println("• No generated files found to clean")
	}
}

// deleteIfExists deletes a file if it exists, returns (deleted, error)
func deleteIfExists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil // File doesn't exist, not an error
	} else if err != nil {
		return false, err // Some other error checking the file
	}

	// File exists, try to delete it
	if err := os.Remove(path); err != nil {
		return false, err
	}

	return true, nil
}

func generateAll(s *scanner.Scanner, cfg *config.Config) {
	if cfg.Generation.Routes.Enabled {
		generateRoutes(s, cfg)
	}
	if cfg.Generation.Dependencies.Enabled {
		generateDeps(s, cfg)
	}

	// Generate Swagger documentation
	generateSwagger(cfg)
}

func generateSwagger(cfg *config.Config) {
	spinner := NewSpinner()
	spinner.Start("Generating Swagger documentation...")

	// Check if swag command is available
	if !isCommandAvailable("swag") {
		spinner.Stop("Installing swag command...")
		installSpinner := NewSpinner()
		installSpinner.Start("Installing swag...")

		if err := installSwag(); err != nil {
			installSpinner.Stop("Failed to install swag")
			fmt.Printf("  Please install manually: go install github.com/swaggo/swag/cmd/swag@latest\n")
			return
		}
		installSpinner.Stop("swag installed successfully")
	}

	// Generate swagger docs
	// Look for main.go in common locations
	mainFile := findMainFile()
	if mainFile == "" {
		spinner.Stop("Could not find main.go file for swagger generation")
		return
	}

	docsDir := "docs"
	cmd := exec.Command("swag", "init", "-g", mainFile, "-o", docsDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		spinner.Stop("Error generating swagger docs")
		fmt.Printf("Output: %s\n", string(output))
		return
	}

	spinner.Stop(fmt.Sprintf("Swagger documentation generated successfully at %s/", docsDir))
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func installSwag() error {
	cmd := exec.Command("go", "install", "github.com/swaggo/swag/cmd/swag@latest")
	return cmd.Run()
}

func findMainFile() string {
	// Common locations for main.go
	candidates := []string{
		"./cmd/server/main.go",
		"./cmd/main.go",
		"./main.go",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

func generateRoutes(s *scanner.Scanner, cfg *config.Config) {
	if !cfg.Generation.Routes.Enabled {
		return
	}

	spinner := NewSpinner()
	spinner.Start("Generating routes...")

	handlers, routes, err := s.ScanRoutes(cfg.Paths.ScanDirs)
	if err != nil {
		spinner.Stop("Error scanning routes")
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(handlers) == 0 {
		spinner.Stop("No handlers found")
		return
	}

	if len(routes) == 0 {
		spinner.Stop("No @Router annotations found")
		return
	}

	// Generate routes using the RouteGenerator
	routeGen := generator.NewRouteGenerator(cfg)
	if err := routeGen.GenerateRoutes(handlers, routes); err != nil {
		spinner.Stop("Error generating routes")
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	outputPath := filepath.Join(cfg.Paths.OutputDir, cfg.Generation.Routes.OutputFile)
	spinner.Stop("Routes generated successfully")
	fmt.Printf("  • Found %d handlers and %d routes\n", len(handlers), len(routes))
	fmt.Printf("  • Generated: %s\n", outputPath)
}

func generateDeps(s *scanner.Scanner, cfg *config.Config) {
	if !cfg.Generation.Dependencies.Enabled {
		return
	}

	spinner := NewSpinner()
	spinner.Start("Generating dependencies...")

	providers, err := s.ScanProviders(cfg.Paths.ScanDirs)
	if err != nil {
		spinner.Stop("Error scanning providers")
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(providers) == 0 {
		spinner.Stop("No provider functions found")
		return
	}

	// Generate dependencies using the DependencyGenerator
	depGen := generator.NewDependencyGenerator(cfg)
	if err := depGen.GenerateDependencies(providers); err != nil {
		spinner.Stop("Error generating dependencies")
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	outputPath := filepath.Join(cfg.Paths.OutputDir, cfg.Generation.Dependencies.OutputFile)
	spinner.Stop("Dependencies generated successfully")
	fmt.Printf("  • Found %d providers\n", len(providers))
	fmt.Printf("  • Generated: %s\n", outputPath)
}
