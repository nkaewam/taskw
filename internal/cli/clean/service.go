package clean

import (
	"os"
	"path/filepath"

	"github.com/nkaewam/taskw/internal/cli/file"
	"github.com/nkaewam/taskw/internal/cli/ui"
	"github.com/nkaewam/taskw/internal/config"
)

// Service handles cleanup of generated files
type Service interface {
	// Clean removes all generated files and reports what was cleaned
	Clean() (deletedFiles []string, skippedFiles []string, err error)
}

// service implements Service interface
type service struct {
	config      *config.Config
	ui          ui.Service
	fileService file.Service
}

// ProvideCleanService creates a new clean service
// @Provider
func ProvideCleanService(config *config.Config, uiService ui.Service, fileService file.Service) Service {
	return &service{
		config:      config,
		ui:          uiService,
		fileService: fileService,
	}
}

// Clean removes all generated files and reports what was cleaned
func (s *service) Clean() ([]string, []string, error) {
	stopSpinner := s.ui.ShowSpinner("Cleaning generated files...")

	var deletedFiles []string
	var skippedFiles []string

	// Clean routes file
	if s.config.Generation.Routes.Enabled {
		routesPath := filepath.Join(s.config.Paths.OutputDir, s.config.Generation.Routes.OutputFile)
		if deleted, err := s.fileService.DeleteIfExists(routesPath); err != nil {
			stopSpinner("Clean completed with errors")
			return deletedFiles, skippedFiles, err
		} else if deleted {
			deletedFiles = append(deletedFiles, routesPath)
		} else {
			skippedFiles = append(skippedFiles, routesPath)
		}
	}

	// Clean dependencies file
	if s.config.Generation.Dependencies.Enabled {
		depsPath := filepath.Join(s.config.Paths.OutputDir, s.config.Generation.Dependencies.OutputFile)
		if deleted, err := s.fileService.DeleteIfExists(depsPath); err != nil {
			stopSpinner("Clean completed with errors")
			return deletedFiles, skippedFiles, err
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
		if deleted, err := s.fileService.DeleteIfExists(swaggerFile); err != nil {
			stopSpinner("Clean completed with errors")
			return deletedFiles, skippedFiles, err
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

	stopSpinner("Clean completed successfully")
	return deletedFiles, skippedFiles, nil
}
