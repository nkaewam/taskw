package e2e

import (
	"os/exec"
	"testing"
)

// getTaskwBinary returns the path to the taskw-dev binary for testing
func getTaskwBinary(t *testing.T) string {
	// Look for taskw-dev in PATH
	if path, err := exec.LookPath("taskw"); err == nil {
		return path
	}

	t.Fatalf("Could not find taskw binary in PATH. Please ensure it is built and available.")
	return ""
}
