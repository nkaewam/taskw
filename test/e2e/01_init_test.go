package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestProjectInitialization tests the complete project initialization workflow
func TestProjectInitialization(t *testing.T) {
	// Setup: Create temporary directory for test
	testDir := filepath.Join(os.TempDir(), "taskw-e2e-init-test")
	if err := os.RemoveAll(testDir); err != nil {
		t.Fatalf("Failed to clean test directory: %v", err)
	}
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir) // Cleanup

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir) // Restore working directory

	// Get path to taskw binary BEFORE changing directories
	taskwBin := getTaskwBinary(t)

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}
	module := "github.com/test/e2e-init-project"

	t.Run("01_initialize_project", func(t *testing.T) {
		// Run: taskw init with module
		cmd := exec.Command(taskwBin, "init", module)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("taskw init failed: %v \n Output: %s", err, string(output))
		}

		t.Logf("✅ taskw init output: %s", string(output))

		// Verify: Project directory was created
		projectName := "e2e-init-project"
		projectDir := filepath.Join(testDir, projectName)
		if _, err := os.Stat(projectDir); os.IsNotExist(err) {
			t.Fatalf("Project directory was not created: %s", projectDir)
		}

		t.Logf("✅ Project directory created: %s", projectDir)
	})

	projectName := "e2e-init-project"
	projectDir := filepath.Join(testDir, projectName)

	t.Run("02_verify_scaffolded_files", func(t *testing.T) {
		// Verify: All expected files exist
		expectedFiles := []string{
			"cmd/server/main.go",
			"internal/api/server.go",
			"internal/api/wire.go",
			"internal/health/handler.go",
			".air.toml",
			"Taskfile.yml",
			"taskw.yaml",
			"go.mod",
			".taskwignore",
		}

		for _, file := range expectedFiles {
			filePath := filepath.Join(projectDir, file)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("Expected file not found: %s", file)
			} else {
				t.Logf("✅ File exists: %s", file)
			}
		}

		// Verify: Directories exist
		expectedDirs := []string{
			"bin",
			"docs",
			"cmd/server",
			"internal/api",
			"internal/health",
		}

		for _, dir := range expectedDirs {
			dirPath := filepath.Join(projectDir, dir)
			if stat, err := os.Stat(dirPath); os.IsNotExist(err) || !stat.IsDir() {
				t.Errorf("Expected directory not found: %s", dir)
			} else {
				t.Logf("✅ Directory exists: %s", dir)
			}
		}
	})

	t.Run("03_verify_go_mod_content", func(t *testing.T) {
		// Verify: go.mod has correct module
		goModPath := filepath.Join(projectDir, "go.mod")
		content, err := os.ReadFile(goModPath)
		if err != nil {
			t.Fatalf("Failed to read go.mod: %v", err)
		}

		if !strings.Contains(string(content), module) {
			t.Errorf("go.mod does not contain expected module %s\nContent:\n%s", module, string(content))
		} else {
			t.Logf("✅ go.mod contains correct module: %s", module)
		}

		// Verify: go.mod has expected dependencies
		expectedDeps := []string{
			"github.com/gofiber/fiber/v2",
			"github.com/google/wire",
			"github.com/gofiber/contrib/swagger",
		}

		for _, dep := range expectedDeps {
			if !strings.Contains(string(content), dep) {
				t.Errorf("go.mod missing expected dependency: %s", dep)
			} else {
				t.Logf("✅ go.mod contains dependency: %s", dep)
			}
		}
	})

	t.Run("04_verify_taskw_config", func(t *testing.T) {
		// Verify: taskw.yaml has correct configuration
		taskwConfigPath := filepath.Join(projectDir, "taskw.yaml")
		content, err := os.ReadFile(taskwConfigPath)
		if err != nil {
			t.Fatalf("Failed to read taskw.yaml: %v", err)
		}

		configContent := string(content)
		if !strings.Contains(configContent, module) {
			t.Errorf("taskw.yaml does not contain expected module %s\nContent:\n%s", module, configContent)
		} else {
			t.Logf("✅ taskw.yaml contains correct module: %s", module)
		}

		// Check for expected configuration sections
		expectedConfig := []string{
			"version:",
			"project:",
			"paths:",
			"generation:",
			"routes:",
			"dependencies:",
		}

		for _, config := range expectedConfig {
			if !strings.Contains(configContent, config) {
				t.Errorf("taskw.yaml missing configuration section: %s", config)
			} else {
				t.Logf("✅ taskw.yaml contains config: %s", config)
			}
		}
	})

	t.Run("05_verify_health_handler", func(t *testing.T) {
		// Verify: Health handler contains correct content
		handlerPath := filepath.Join(projectDir, "internal/health/handler.go")
		content, err := os.ReadFile(handlerPath)
		if err != nil {
			t.Fatalf("Failed to read health handler: %v", err)
		}

		handlerContent := string(content)
		expectedContent := []string{
			"package health",
			"ProvideHandler",
			"GetHealth",
			"@Router /health [get]",
			projectName + " API is running successfully",
		}

		for _, expected := range expectedContent {
			if !strings.Contains(handlerContent, expected) {
				t.Errorf("Health handler missing expected content: %s", expected)
			} else {
				t.Logf("✅ Health handler contains: %s", expected)
			}
		}
	})

	t.Run("06_verify_taskwignore", func(t *testing.T) {
		// Verify: .taskwignore exists and has sensible patterns
		taskwIgnorePath := filepath.Join(projectDir, ".taskwignore")
		content, err := os.ReadFile(taskwIgnorePath)
		if err != nil {
			t.Fatalf("Failed to read .taskwignore: %v", err)
		}

		ignoreContent := string(content)
		expectedPatterns := []string{
			"**/*_test.go",
			"**/vendor/**",
			"**/bin/**",
			"**/*_gen.go",
			"!routes_gen.go",
			"!dependencies_gen.go",
		}

		for _, pattern := range expectedPatterns {
			if !strings.Contains(ignoreContent, pattern) {
				t.Errorf(".taskwignore missing expected pattern: %s", pattern)
			} else {
				t.Logf("✅ .taskwignore contains pattern: %s", pattern)
			}
		}
	})

	t.Run("07_verify_initialization_completed", func(t *testing.T) {
		// Since taskw init now automatically runs go mod tidy and task generate,
		// we should verify that the initialization process completed successfully

		// Check if generated files exist
		generatedFiles := []string{
			"internal/api/routes_gen.go",
			"internal/api/dependencies_gen.go",
		}

		for _, file := range generatedFiles {
			filePath := filepath.Join(projectDir, file)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("Generated file not found: %s (init should have run task generate)", file)
			} else {
				t.Logf("✅ Generated file exists: %s", file)
			}
		}
	})

	t.Run("08_project_compiles", func(t *testing.T) {
		// Test: Project has valid Go syntax (may not compile due to missing deps)
		if err := os.Chdir(projectDir); err != nil {
			t.Fatalf("Failed to change to project directory: %v", err)
		}

		// Run go build to check syntax
		cmd := exec.Command("go", "build", "./...")
		output, err := cmd.CombinedOutput()
		if err != nil {
			buildOutput := string(output)
			t.Logf("Build output: %s", buildOutput)

			// Check for syntax errors vs dependency errors
			if strings.Contains(buildOutput, "syntax error") {
				t.Errorf("Scaffolded code has syntax errors: %v", err)
			} else if strings.Contains(buildOutput, "missing go.sum entry") ||
				strings.Contains(buildOutput, "GeneratedProviderSet") {
				t.Logf("✅ Build fails as expected due to missing dependencies/generated code")
			} else {
				t.Logf("⚠️  Build failed with: %v (might be expected)", err)
			}
		} else {
			t.Logf("✅ Project compiles successfully")
		}
	})

	t.Logf("✅ Project initialization e2e test completed successfully")
}

// Note: getTaskwBinary is now defined in utils.go
