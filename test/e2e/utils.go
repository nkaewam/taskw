package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// getTaskwBinary returns the path to the taskw binary for testing
func getTaskwBinary(t *testing.T) string {
	// First try to find the binary relative to the project root
	// Get current working directory and find project root
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Look for project root by finding go.mod with taskw module
	projectRoot := cwd
	maxDepth := 10 // Prevent infinite loop
	for i := 0; i < maxDepth; i++ {
		goModPath := filepath.Join(projectRoot, "go.mod")
		if content, err := os.ReadFile(goModPath); err == nil {
			// Check if this is the taskw project
			if strings.Contains(string(content), "github.com/nkaewam/taskw") {
				break
			}
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			// Reached filesystem root without finding taskw project
			projectRoot = ""
			break
		}
		projectRoot = parent
	}

	// Try the binary in the project root if found
	if projectRoot != "" {
		binaryPath := filepath.Join(projectRoot, "bin", "taskw")
		if _, err := os.Stat(binaryPath); err == nil {
			return binaryPath
		}
	}

	// Try common relative paths from test directory
	possiblePaths := []string{
		"./bin/taskw",
		"../bin/taskw",
		"../../bin/taskw",
		"../../../bin/taskw", // From test/e2e directory
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, err := filepath.Abs(path)
			if err != nil {
				t.Fatalf("Failed to get absolute path for taskw binary: %v", err)
			}
			return absPath
		}
	}

	// Try to find taskw in PATH
	if path, err := exec.LookPath("taskw"); err == nil {
		return path
	}

	t.Fatalf("Could not find taskw binary. Please build it first with: go build -o bin/taskw cmd/taskw/main.go\nLooked in project root: %s\nCurrent working dir: %s", projectRoot, cwd)
	return ""
}
