package generator

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/init
var initTemplateFS embed.FS

// InitGenerator creates new projects from templates
type InitGenerator struct{}

// NewInitGenerator creates a new init generator
func NewInitGenerator() *InitGenerator {
	return &InitGenerator{}
}

// InitProject scaffolds a new project with the specified configuration
func (g *InitGenerator) InitProject(projectPath, module, projectName string) error {
	// Create project directory if it doesn't exist
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Template data
	data := struct {
		Module      string
		ProjectName string
		BinaryName  string
	}{
		Module:      module,
		ProjectName: projectName,
		BinaryName:  strings.ReplaceAll(strings.ToLower(projectName), " ", "-"),
	}

	// Files to create with their templates
	files := []struct {
		template string
		output   string
	}{
		{"templates/init/cmd/server/main.tmpl", "cmd/server/main.go"},
		{"templates/init/internal/api/server.tmpl", "internal/api/server.go"},
		{"templates/init/internal/api/wire.tmpl", "internal/api/wire.go"},
		{"templates/init/internal/health/handler.tmpl", "internal/health/handler.go"},
		{"templates/init/docs/docs.tmpl", "docs/docs.go"},
		{"templates/init/air.tmpl", ".air.toml"},
		{"templates/init/Taskfile.tmpl", "Taskfile.yml"},
		{"templates/init/taskw.tmpl", "taskw.yaml"},
		{"templates/init/go_mod.tmpl", "go.mod"},
	}

	// Generate each file
	for _, file := range files {
		if err := g.generateFile(projectPath, file.template, file.output, data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", file.output, err)
		}
		fmt.Printf("Created: %s\n", file.output)
	}

	// Create additional directories
	directories := []string{
		"bin",
		"docs",
	}

	for _, dir := range directories {
		dirPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		fmt.Printf("Created directory: %s/\n", dir)
	}

	// Create or append to .taskwignore
	if err := g.createOrAppendTaskwIgnore(projectPath); err != nil {
		fmt.Printf("Warning: Failed to create/update .taskwignore: %v\n", err)
	}

	// Automatically generate code after scaffolding
	if err := g.runInitialGeneration(projectPath); err != nil {
		// Don't fail the entire init process, just warn the user
		return fmt.Errorf("warning: Failed to run initial code generation: %v", err)
	}

	return nil
}

// generateFile generates a single file from a template
func (g *InitGenerator) generateFile(projectPath, templatePath, outputPath string, data interface{}) error {
	// Read template
	tmplContent, err := initTemplateFS.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	// Parse template
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	// Create output directory
	outputDir := filepath.Dir(filepath.Join(projectPath, outputPath))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// Generate content
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	// Write file
	outputFile := filepath.Join(projectPath, outputPath)
	if err := os.WriteFile(outputFile, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outputFile, err)
	}

	return nil
}

// runInitialGeneration runs go mod tidy and then task generate in the newly created project
func (g *InitGenerator) runInitialGeneration(projectPath string) error {
	// Check if go command is available
	if !isCommandAvailable("go") {
		return fmt.Errorf("go command not available in PATH, bro what?")
	}

	// Check if task command is available
	if !isCommandAvailable("task") {
		return fmt.Errorf("task command not available, please install Task runner or run 'go install github.com/go-task/task/v3/cmd/task@latest'")
	}

	// Step 1: Run go mod tidy to resolve dependencies
	fmt.Println("📦 Running go mod tidy to resolve dependencies...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = projectPath

	tidyOutput, err := tidyCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run 'go mod tidy': %w\nOutput: %s", err, string(tidyOutput))
	}
	fmt.Println("✅ Dependencies resolved successfully")

	// Step 2: Run task generate directly to create initial code
	fmt.Println("🔧 Running task generate to create initial code...")
	generateCmd := exec.Command("task", "generate")
	generateCmd.Dir = projectPath

	// Capture output for better error reporting
	generateOutput, err := generateCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run 'task generate': %w\nOutput: %s", err, string(generateOutput))
	}

	return nil
}

// isCommandAvailable checks if a command is available in PATH
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// createOrAppendTaskwIgnore creates or appends to .taskwignore file
func (g *InitGenerator) createOrAppendTaskwIgnore(projectPath string) error {
	taskwIgnorePath := filepath.Join(projectPath, ".taskwignore")

	// Try to read existing .taskwignore as template (optional)
	templateContent, usingTemplate := g.readCurrentTaskwIgnore()
	if !usingTemplate {
		// No .taskwignore found in current project, use default patterns
		templateContent = g.getDefaultTaskwIgnoreContent()
	}

	// Check if .taskwignore already exists in the target project
	if _, err := os.Stat(taskwIgnorePath); err == nil {
		// File exists, append our content
		existingContent, err := os.ReadFile(taskwIgnorePath)
		if err != nil {
			return fmt.Errorf("failed to read existing .taskwignore: %w", err)
		}

		// Check if our content is already present
		if strings.Contains(string(existingContent), "# Taskw Ignore Patterns") {
			fmt.Printf("Updated: .taskwignore (Taskw patterns already present)\n")
			return nil
		}

		// Append our content with a separator
		separator := "\n\n# ===== Taskw Default Patterns =====\n"
		combinedContent := string(existingContent) + separator + templateContent

		if err := os.WriteFile(taskwIgnorePath, []byte(combinedContent), 0644); err != nil {
			return fmt.Errorf("failed to append to .taskwignore: %w", err)
		}

		fmt.Printf("Updated: .taskwignore (appended Taskw patterns)\n")
	} else {
		// File doesn't exist, create new one
		if err := os.WriteFile(taskwIgnorePath, []byte(templateContent), 0644); err != nil {
			return fmt.Errorf("failed to create .taskwignore: %w", err)
		}

		if usingTemplate {
			fmt.Printf("Created: .taskwignore (using patterns from current project)\n")
		} else {
			fmt.Printf("Created: .taskwignore (using default patterns)\n")
		}
	}

	return nil
}

// readCurrentTaskwIgnore reads the .taskwignore from the current working directory
// Returns (content, found) where found indicates if a .taskwignore was found
func (g *InitGenerator) readCurrentTaskwIgnore() (string, bool) {
	// Try to find .taskwignore in current directory first
	possiblePaths := []string{
		".taskwignore",
		"../.taskwignore", // Check parent directory (useful when running from subdirectory)
	}

	for _, path := range possiblePaths {
		if content, err := os.ReadFile(path); err == nil {
			return string(content), true
		}
	}

	// .taskwignore is optional - return empty content and false
	return "", false
}

// getDefaultTaskwIgnoreContent returns fallback .taskwignore content
func (g *InitGenerator) getDefaultTaskwIgnoreContent() string {
	return `# Taskw Ignore Patterns
# This file tells taskw which files and directories to exclude from scanning

# Test files
**/*_test.go
**/testdata/**
**/test/**
**/*_mock.go

# Build artifacts
**/bin/**
**/build/**
**/dist/**
target/

# Dependencies
**/vendor/**
**/node_modules/**

# Generated files (except the ones taskw generates)
**/*_gen.go
!routes_gen.go
!dependencies_gen.go
**/wire_gen.go

# IDE and editor files
.vscode/
.idea/
**/*.swp
**/*.swo
**/*~

# OS files
.DS_Store
Thumbs.db

# Temporary files
**/*.tmp
**/*.temp
**/*.log

# Git
.git/
.gitignore

# Documentation (optional)
*.md
!README.md

# Configuration files (optional)
*.yaml
*.yml
*.json
*.toml
!taskw.yaml

# Main/cmd files that don't contain handlers
cmd/
main.go

# Models and shared types (no handlers here)
**/models/**
**/types/**
**/errors/**
**/utils/**
**/config/**
`
}

// ValidateProjectPath checks if the project path is valid for initialization
func (g *InitGenerator) ValidateProjectPath(projectPath string) error {
	// Check if directory exists and is empty
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		// Directory doesn't exist - that's fine, we'll create it
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("cannot read directory: %w", err)
	}

	// Check if directory is empty (ignore hidden files like .git, .DS_Store)
	nonHiddenFiles := 0
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".") {
			nonHiddenFiles++
		}
	}

	if nonHiddenFiles > 0 {
		return fmt.Errorf("directory %s is not empty (contains %d non-hidden files)", projectPath, nonHiddenFiles)
	}

	return nil
}
