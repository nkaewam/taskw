package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nkaewam/taskw/internal/config"
	"github.com/nkaewam/taskw/internal/scanner"
)

func main() {
	var (
		configPath = flag.String("config", "", "Path to taskw.yaml config file")
		help       = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Error: command required")
		showUsage()
		os.Exit(1)
	}

	command := args[0]

	switch command {
	case "init":
		handleInit(*configPath)
	case "generate":
		handleGenerate(args[1:], *configPath)
	case "scan":
		handleScan(*configPath)
	default:
		fmt.Printf("Error: unknown command '%s'\n", command)
		showUsage()
		os.Exit(1)
	}
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

		fmt.Println("✅ Created taskw.yaml")
		fmt.Printf("📁 Configured to scan: %v\n", cfg.Paths.ScanDirs)
		fmt.Printf("📂 Output directory: %s\n", cfg.Paths.OutputDir)
	}

	// Create .taskwignore if it doesn't exist
	if _, err := os.Stat(".taskwignore"); os.IsNotExist(err) {
		filter := scanner.NewFileFilter()
		if err := filter.CreateDefaultTaskwIgnore(); err != nil {
			fmt.Printf("Warning: Could not create .taskwignore: %v\n", err)
		} else {
			fmt.Println("✅ Created .taskwignore")
		}
	} else {
		fmt.Println("📋 Using existing .taskwignore")
	}

	fmt.Println("\n📝 Next steps:")
	fmt.Println("  1. Edit taskw.yaml to configure scan directories and output settings")
	fmt.Println("  2. Edit .taskwignore to exclude files/directories from scanning")
	fmt.Println("  3. Run 'taskw scan' to preview what will be generated")
	fmt.Println("  4. Run 'taskw generate' to generate code")
}

func handleGenerate(args []string, configPath string) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	target := "all"
	if len(args) > 0 {
		target = args[0]
	}

	// Create scanner
	s := scanner.NewScanner(cfg)

	switch target {
	case "all":
		generateAll(s, cfg)
	case "routes":
		generateRoutes(s, cfg)
	case "deps", "dependencies":
		generateDeps(s, cfg)
	default:
		fmt.Printf("Error: unknown generate target '%s'\n", target)
		os.Exit(1)
	}
}

func handleScan(configPath string) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create scanner
	s := scanner.NewScanner(cfg)

	fmt.Println("🔍 Scanning codebase...")
	fmt.Println("📋 Using ignore patterns from .taskwignore")

	// Scan all configured directories
	result, err := s.ScanAll()
	if err != nil {
		fmt.Printf("Error scanning: %v\n", err)
		os.Exit(1)
	}

	// Display results
	stats := s.GetStatistics(result)
	fmt.Printf("\n📊 Scan Results:\n")
	fmt.Printf("   🎯 Handlers found: %d\n", stats.HandlersFound)
	fmt.Printf("   🛣️  Routes found: %d\n", stats.RoutesFound)
	fmt.Printf("   📦 Providers found: %d\n", stats.ProvidersFound)
	fmt.Printf("   📄 Packages scanned: %d\n", stats.PackagesScanned)

	if stats.ErrorsFound > 0 {
		fmt.Printf("   ❌ Errors: %d\n", stats.ErrorsFound)
	}

	// Show detailed results if requested
	if len(result.Handlers) > 0 {
		fmt.Println("\n🎯 Handlers:")
		for _, h := range result.Handlers {
			fmt.Printf("   - %s.%s (%s)\n", h.Package, h.FunctionName, h.HandlerName)
		}
	}

	if len(result.Routes) > 0 {
		fmt.Println("\n🛣️  Routes:")
		for _, r := range result.Routes {
			fmt.Printf("   - %s %s -> %s\n", r.HTTPMethod, r.Path, r.HandlerRef)
		}
	}

	if len(result.Providers) > 0 {
		fmt.Println("\n📦 Providers:")
		for _, p := range result.Providers {
			fmt.Printf("   - %s() -> %s\n", p.FunctionName, p.ReturnType)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Println("\n❌ Errors:")
		for _, e := range result.Errors {
			fmt.Printf("   - %s: %s\n", e.FilePath, e.Message)
		}
	}

	// Validate results
	validator := scanner.NewValidator()
	validation := validator.ValidateScanResult(result)

	if validation.HasErrors() {
		fmt.Println("\n🚨 Validation Errors:")
		for _, err := range validation.Errors {
			fmt.Printf("   - %s: %s\n", err.Type, err.Message)
		}
	}

	if validation.HasWarnings() {
		fmt.Println("\n⚠️  Validation Warnings:")
		for _, warn := range validation.Warnings {
			fmt.Printf("   - %s: %s\n", warn.Type, warn.Message)
		}
	}
}

func generateAll(s *scanner.Scanner, cfg *config.Config) {
	if cfg.Generation.Routes.Enabled {
		generateRoutes(s, cfg)
	}
	if cfg.Generation.Dependencies.Enabled {
		generateDeps(s, cfg)
	}
}

func generateRoutes(s *scanner.Scanner, cfg *config.Config) {
	if !cfg.Generation.Routes.Enabled {
		return
	}

	fmt.Println("🔄 Generating routes...")

	handlers, routes, err := s.ScanRoutes(cfg.Paths.ScanDirs)
	if err != nil {
		fmt.Printf("Error scanning routes: %v\n", err)
		os.Exit(1)
	}

	if len(handlers) == 0 {
		fmt.Println("⚠️  No handlers found")
		return
	}

	if len(routes) == 0 {
		fmt.Println("⚠️  No @Router annotations found")
		return
	}

	outputPath := filepath.Join(cfg.Paths.OutputDir, cfg.Generation.Routes.OutputFile)
	fmt.Printf("✅ Found %d handlers and %d routes\n", len(handlers), len(routes))
	fmt.Printf("📝 Route generation will write to: %s\n", outputPath)

	// TODO: Implement actual route generation with templates
	fmt.Println("⚠️  Route generation not implemented yet - this will be added in the next step")
}

func generateDeps(s *scanner.Scanner, cfg *config.Config) {
	if !cfg.Generation.Dependencies.Enabled {
		return
	}

	fmt.Println("🔄 Generating dependencies...")

	providers, err := s.ScanProviders(cfg.Paths.ScanDirs)
	if err != nil {
		fmt.Printf("Error scanning providers: %v\n", err)
		os.Exit(1)
	}

	if len(providers) == 0 {
		fmt.Println("⚠️  No provider functions found")
		return
	}

	outputPath := filepath.Join(cfg.Paths.OutputDir, cfg.Generation.Dependencies.OutputFile)
	fmt.Printf("✅ Found %d providers\n", len(providers))
	fmt.Printf("📝 Dependency generation will write to: %s\n", outputPath)

	// TODO: Implement actual dependency generation with templates
	fmt.Println("⚠️  Dependency generation not implemented yet - this will be added in the next step")
}

func showHelp() {
	fmt.Println(`taskw - Go API Code Generator (Fiber + Wire + Swaggo)

Usage:
  taskw init                    Create taskw.yaml config and .taskwignore
  taskw generate [target]       Generate code (all, routes, deps)
  taskw scan                    Show what will be generated

Flags:
  --config path                 Path to config file (default: taskw.yaml)
  --help                        Show this help

Examples:
  taskw init                    # Initialize new project
  taskw generate                # Generate everything
  taskw generate routes         # Generate only routes
  taskw generate deps           # Generate only dependencies  
  taskw scan                    # Preview what will be generated

Targets:
  all                          Generate routes and dependencies (default)
  routes                       Generate Fiber route registration
  deps, dependencies           Generate Wire dependency injection

File Filtering:
  TaskW scans all *.go files except those matching patterns in .taskwignore
  The .taskwignore file uses gitignore-style glob patterns:
  
    vendor/**           # Ignore entire vendor directory
    **/*_test.go        # Ignore all test files
    build/              # Ignore build directory
    *.tmp               # Ignore temporary files
  
  Default excludes: vendor/, node_modules/, .git/, build/, test files`)
}

func showUsage() {
	fmt.Println("Use 'taskw --help' for usage information")
}
