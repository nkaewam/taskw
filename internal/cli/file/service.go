package file

import (
	"os"
	"os/exec"
)

// Service handles file system operations
type Service interface {
	// DeleteIfExists deletes a file if it exists, returns (deleted, error)
	DeleteIfExists(path string) (bool, error)
	// IsCommandAvailable checks if a command is available in PATH
	IsCommandAvailable(name string) bool
	// InstallSwag installs the swag command for swagger generation
	InstallSwag() error
	// FindMainFile finds the main.go file in common locations
	FindMainFile() string
}

// service implements Service interface
type service struct{}

// ProvideFileService creates a new file service
// @Provider
func ProvideFileService() Service {
	return &service{}
}

// DeleteIfExists deletes a file if it exists, returns (deleted, error)
func (s *service) DeleteIfExists(path string) (bool, error) {
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

// IsCommandAvailable checks if a command is available in PATH
func (s *service) IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// InstallSwag installs the swag command for swagger generation
func (s *service) InstallSwag() error {
	cmd := exec.Command("go", "install", "github.com/swaggo/swag/cmd/swag@latest")
	return cmd.Run()
}

// FindMainFile finds the main.go file in common locations
func (s *service) FindMainFile() string {
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
