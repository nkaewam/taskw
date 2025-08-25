package taskw

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nkaewam/taskw/internal/cli"
	"github.com/spf13/cobra"
)

var (
	configPath string
	container  *cli.Container
)

var rootCmd = &cobra.Command{
	Use:   "taskw",
	Short: "Go API Code Generator (Fiber + Wire + Swaggo)",
	Long: `Taskw is a code generator for Go APIs that automatically generates:
- Fiber route registration from handler functions
- Wire dependency injection setup from provider functions  
- Swagger documentation

It scans your Go code for special annotations and generates boilerplate code to wire everything together.`,
	PersistentPreRunE: initializeContainer,
}

func initializeContainer(cmd *cobra.Command, args []string) error {
	var err error
	container, err = cli.InitializeContainer(configPath)
	if err != nil {
		return fmt.Errorf("failed to initialize container: %w", err)
	}
	return nil
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

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	Execute()
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
  taskw init github.com/user/my-api             # Create project with specified module`,
	RunE: handleInit,
}

func handleInit(cmd *cobra.Command, args []string) error {

	// Full project scaffolding
	var module string
	if len(args) == 0 {
		// Interactive prompt for module
		var err error
		module, err = container.UI.PromptForModule()
		if err != nil {
			return fmt.Errorf("failed to get module: %w", err)
		}
	} else {
		module = args[0]
		// Validate module format
		if err := container.Project.ValidateModule(module); err != nil {
			return fmt.Errorf("invalid module: %w", err)
		}
	}

	// Extract project name from module path
	projectName := container.Project.ExtractProjectName(module)
	projectPath := filepath.Join(".", projectName)

	// Validate project directory
	if err := container.Project.ValidateProjectPath(projectPath); err != nil {
		return fmt.Errorf("invalid project path: %w", err)
	}

	stopSpinner := container.UI.ShowSpinner(fmt.Sprintf("Creating project %s...", projectName))

	// Generate the project
	if err := container.Project.InitProject(projectPath, module, projectName); err != nil {
		stopSpinner("Project creation failed")
		return fmt.Errorf("failed to create project: %w", err)
	}

	stopSpinner(fmt.Sprintf("Project %s created successfully", projectName))
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return container.Generation.GenerateAll()
	},
}

var generateRoutesCmd = &cobra.Command{
	Use:   "routes",
	Short: "Generate Fiber route registration",
	Long:  `Generate route registration code from handler functions with @Router annotations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return container.Generation.GenerateRoutes()
	},
}

var generateDepsCmd = &cobra.Command{
	Use:     "deps",
	Aliases: []string{"dependencies"},
	Short:   "Generate Wire dependency injection",
	Long:    `Generate Wire dependency injection setup from provider functions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return container.Generation.GenerateDependencies()
	},
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Show what will be generated",
	Long: `Scan the codebase and display what handlers, routes, and providers would be generated.
This is useful for previewing changes before running generate.`,
	RunE: handleScan,
}

func handleScan(cmd *cobra.Command, args []string) error {
	// Scan all configured directories
	result, err := container.Scan.ScanAll()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Display results
	if err := container.Scan.ShowScanResults(result); err != nil {
		return fmt.Errorf("failed to show results: %w", err)
	}

	// Validate results
	return container.Scan.ValidateScanResults(result)
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove all generated files",
	Long: `Remove all files that were generated by Taskw:
- Route registration files
- Dependency injection files  
- Swagger documentation files

This helps clean up the workspace when regenerating code or switching configurations.`,
	RunE: handleClean,
}

func handleClean(cmd *cobra.Command, args []string) error {
	deletedFiles, skippedFiles, err := container.Clean.Clean()
	if err != nil {
		return fmt.Errorf("clean failed: %w", err)
	}

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

	return nil
}
