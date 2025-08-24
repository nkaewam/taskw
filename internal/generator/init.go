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

	// Automatically generate code after scaffolding
	if err := g.runInitialGeneration(projectPath); err != nil {
		// Don't fail the entire init process, just warn the user
		fmt.Printf("Warning: Failed to run initial code generation: %v\n", err)
		fmt.Println("You can run 'task generate' manually after 'go mod tidy'")
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

// runInitialGeneration runs taskw generate in the newly created project
func (g *InitGenerator) runInitialGeneration(projectPath string) error {
	// Check if task command is available
	if !isCommandAvailable("task") {
		return fmt.Errorf("task command not available, please install Task runner")
	}

	// Check if taskw command is available
	if !isCommandAvailable("taskw") {
		return fmt.Errorf("taskw command not available in PATH")
	}

	// Change to project directory and run task generate
	cmd := exec.Command("task", "generate")
	cmd.Dir = projectPath

	// Capture output for better error reporting
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run 'task generate': %w\nOutput: %s", err, string(output))
	}

	fmt.Println("✅ Code generation completed successfully")
	return nil
}

// isCommandAvailable checks if a command is available in PATH
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
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
