Great idea! Let's scope down to a focused v1 that solves the core problem without over-engineering. This makes it much more achievable and gets you to a working tool faster.

## `taskw` v1 Scope

**Fixed Stack:**

- **DI System**: Wire only
- **Web Framework**: Fiber only
- **Annotations**: Swaggo only
- **Language**: Go only

This removes all the abstraction complexity and focuses on your proven working system.

## Simplified Project Structure

```
taskw/
├── cmd/
│   └── taskw/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── scanner/
│   │   ├── routes.go          # Extract @Router annotations + handlers
│   │   └── providers.go       # Extract Provide* functions
│   └── generator/
│       ├── routes.go          # Generate Fiber routes
│       └── dependencies.go    # Generate Wire dependencies
├── templates/
│   ├── routes.tmpl
│   └── dependencies.tmpl
└── go.mod
```

## Simplified Configuration (`taskw.yaml`)

```yaml
# taskw.yaml - Simple v1 configuration
version: "1.0"

project:
  module: "github.com/user/my-api" # Go module name

paths:
  scan_dirs: # Directories to scan for handlers/providers
    - "./internal/api"
    - "./internal/handlers"
  output_dir: "./internal/api" # Where to generate files

generation:
  routes:
    enabled: true
    output_file: "routes_gen.go"
  dependencies:
    enabled: true
    output_file: "dependencies_gen.go"
```

## CLI Commands (v1)

```bash
# Initialize taskw in existing project
taskw init

# Generate all code
taskw generate

# Generate specific parts
taskw generate routes
taskw generate deps

# Watch mode for development
taskw watch

# Validate/debug
taskw scan     # Show what it found
```

## Implementation Plan

Let me create the core files for v1:
Ran tool

Here's the implementation plan for `taskw` v1. Let me extract the core logic from your existing tools and create a focused CLI:

## 1. Core Configuration

```go
// internal/config/config.go
package config

import (
    "fmt"
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
)

type Config struct {
    Version string  `yaml:"version"`
    Project Project `yaml:"project"`
    Paths   Paths   `yaml:"paths"`
    Generation Generation `yaml:"generation"`
}

type Project struct {
    Module string `yaml:"module"` // Go module name from go.mod
}

type Paths struct {
    ScanDirs  []string `yaml:"scan_dirs"`
    OutputDir string   `yaml:"output_dir"`
}

type Generation struct {
    Routes       RouteConfig `yaml:"routes"`
    Dependencies DepConfig   `yaml:"dependencies"`
}

type RouteConfig struct {
    Enabled    bool   `yaml:"enabled"`
    OutputFile string `yaml:"output_file"`
}

type DepConfig struct {
    Enabled    bool   `yaml:"enabled"`
    OutputFile string `yaml:"output_file"`
}

// LoadConfig loads taskw.yaml or creates default config
func LoadConfig(path string) (*Config, error) {
    if path == "" {
        path = "taskw.yaml"
    }

    // If config doesn't exist, create default
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return createDefaultConfig()
    }

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    return &config, nil
}

func createDefaultConfig() (*Config, error) {
    // Auto-detect Go module
    module, err := detectGoModule()
    if err != nil {
        return nil, err
    }

    return &Config{
        Version: "1.0",
        Project: Project{Module: module},
        Paths: Paths{
            ScanDirs:  []string{"./internal/api", "./internal/handlers"},
            OutputDir: "./internal/api",
        },
        Generation: Generation{
            Routes: RouteConfig{
                Enabled:    true,
                OutputFile: "routes_gen.go",
            },
            Dependencies: DepConfig{
                Enabled:    true,
                OutputFile: "dependencies_gen.go",
            },
        },
    }, nil
}

func detectGoModule() (string, error) {
    data, err := os.ReadFile("go.mod")
    if err != nil {
        return "", err
    }

    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "module ") {
            return strings.TrimSpace(line[7:]), nil
        }
    }

    return "", fmt.Errorf("could not detect Go module name")
}

func (c *Config) Save(path string) error {
    if path == "" {
        path = "taskw.yaml"
    }

    data, err := yaml.Marshal(c)
    if err != nil {
        return err
    }

    return os.WriteFile(path, data, 0644)
}
```

## 2. Scanner Logic (Extracted from your tools)

```go
// internal/scanner/types.go
package scanner

type HandlerFunction struct {
    FunctionName string
    Package      string
    ReturnType   string
    HandlerName  string
}

type RouteMapping struct {
    MethodName string
    Path       string
    HTTPMethod string
}

type ProviderFunction struct {
    FunctionName string
    Package      string
    ReturnType   string
    FileType     string
}
```

```go
// internal/scanner/routes.go - Based on your generate-routes/main.go
package scanner

import (
    "bufio"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

func ScanRoutes(directories []string) ([]HandlerFunction, []RouteMapping, error) {
    var allHandlers []HandlerFunction
    var allRouteMappings []RouteMapping

    for _, dir := range directories {
        handlers, mappings, err := scanDirectory(dir)
        if err != nil {
            return nil, nil, err
        }
        allHandlers = append(allHandlers, handlers...)
        allRouteMappings = append(allRouteMappings, mappings...)
    }

    return allHandlers, allRouteMappings, nil
}

func scanDirectory(dir string) ([]HandlerFunction, []RouteMapping, error) {
    var allHandlers []HandlerFunction
    var allRouteMappings []RouteMapping

    err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Only process handler.go files
        if !info.IsDir() && strings.HasSuffix(filepath.Base(path), "handler.go") {
            handlers, err := extractHandlersFromFile(path)
            if err != nil {
                return err
            }
            allHandlers = append(allHandlers, handlers...)

            routeMappings, err := extractRoutesFromFile(path)
            if err != nil {
                return err
            }
            allRouteMappings = append(allRouteMappings, routeMappings...)
        }

        return nil
    })

    return allHandlers, allRouteMappings, err
}

// Copy your existing extractHandlersFromFile and extractRoutesFromFile functions here
// with minimal modifications
```

## 3. Simple CLI Structure

```go
// cmd/taskw/main.go
package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/user/taskw/internal/config"
    "github.com/user/taskw/internal/generator"
    "github.com/user/taskw/internal/scanner"
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
        handleInit()
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

func handleInit() {
    cfg, err := config.LoadConfig("")
    if err != nil {
        fmt.Printf("Error creating config: %v\n", err)
        os.Exit(1)
    }

    if err := cfg.Save("taskw.yaml"); err != nil {
        fmt.Printf("Error saving config: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("✅ Created taskw.yaml")
    fmt.Println("Edit the config and run 'taskw generate' to generate code")
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

    switch target {
    case "all":
        generateRoutes(cfg)
        generateDeps(cfg)
    case "routes":
        generateRoutes(cfg)
    case "deps":
        generateDeps(cfg)
    default:
        fmt.Printf("Error: unknown generate target '%s'\n", target)
        os.Exit(1)
    }
}

func generateRoutes(cfg *config.Config) {
    if !cfg.Generation.Routes.Enabled {
        return
    }

    fmt.Println("🔄 Generating routes...")

    handlers, mappings, err := scanner.ScanRoutes(cfg.Paths.ScanDirs)
    if err != nil {
        fmt.Printf("Error scanning routes: %v\n", err)
        os.Exit(1)
    }

    outputPath := filepath.Join(cfg.Paths.OutputDir, cfg.Generation.Routes.OutputFile)
    if err := generator.GenerateRoutes(handlers, mappings, outputPath, cfg.Project.Module); err != nil {
        fmt.Printf("Error generating routes: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("✅ Generated routes in %s\n", outputPath)
}

func generateDeps(cfg *config.Config) {
    if !cfg.Generation.Dependencies.Enabled {
        return
    }

    fmt.Println("🔄 Generating dependencies...")

    providers, err := scanner.ScanProviders(cfg.Paths.ScanDirs)
    if err != nil {
        fmt.Printf("Error scanning providers: %v\n", err)
        os.Exit(1)
    }

    outputPath := filepath.Join(cfg.Paths.OutputDir, cfg.Generation.Dependencies.OutputFile)
    if err := generator.GenerateDependencies(providers, outputPath, cfg.Project.Module); err != nil {
        fmt.Printf("Error generating dependencies: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("✅ Generated dependencies in %s\n", outputPath)
}

func showHelp() {
    fmt.Println(`taskw - Go API Code Generator (Fiber + Wire + Swaggo)

Usage:
  taskw init                    Create taskw.yaml config
  taskw generate [target]       Generate code (all, routes, deps)
  taskw scan                    Show what will be generated

Flags:
  --config path                 Path to config file (default: taskw.yaml)
  --help                        Show this help

Examples:
  taskw init                    # Initialize new project
  taskw generate                # Generate everything
  taskw generate routes         # Generate only routes
  taskw generate deps           # Generate only dependencies`)
}

func showUsage() {
    fmt.Println("Use 'taskw --help' for usage information")
}
```

## 4. Example Usage

```bash
# 1. In a new Go project
cd my-fiber-api
taskw init

# 2. Edit taskw.yaml to match your structure
# 3. Generate code
taskw generate

# 4. Integration with existing build tools
# Add to Taskfile.yml:
# generate:
#   cmds:
#     - taskw generate
```
