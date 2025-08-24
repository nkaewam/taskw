package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	Long: `TaskW is a code generator for Go APIs that automatically generates:
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
	Use:   "init",
	Short: "Create taskw.yaml config and .taskwignore",
	Long: `Initialize a new TaskW project by creating:
- taskw.yaml configuration file with default settings
- .taskwignore file with common exclusion patterns

This sets up the necessary configuration for TaskW to scan your codebase and generate code.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleInit(configPath)
	},
}

func handleInit(configPath string) {
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
	Long: `Remove all files that were generated by TaskW:
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
