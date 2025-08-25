package scan

import (
	"fmt"
	"strings"

	"github.com/nkaewam/taskw/internal/cli/ui"
	"github.com/nkaewam/taskw/internal/config"
	"github.com/nkaewam/taskw/internal/scanner"
)

// Service handles codebase scanning operations
type Service interface {
	// ScanAll scans all configured directories and returns scan results
	ScanAll() (*scanner.ScanResult, error)
	// ShowScanResults displays scan results to the user
	ShowScanResults(result *scanner.ScanResult) error
	// ValidateScanResults performs validation on scan results
	ValidateScanResults(result *scanner.ScanResult) error
}

// service implements Service interface
type service struct {
	config  *config.Config
	scanner *scanner.Scanner
	ui      ui.Service
}

// ProvideScanService creates a new scan service
// @Provider
func ProvideScanService(config *config.Config, uiService ui.Service) Service {
	return &service{
		config:  config,
		scanner: scanner.NewScanner(config),
		ui:      uiService,
	}
}

// ScanAll scans all configured directories and returns scan results
func (s *service) ScanAll() (*scanner.ScanResult, error) {
	stopSpinner := s.ui.ShowSpinner("Scanning codebase...")
	fmt.Println("• Using ignore patterns from .taskwignore")

	result, err := s.scanner.ScanAll()
	if err != nil {
		stopSpinner("Scan failed")
		return nil, fmt.Errorf("error scanning: %w", err)
	}

	stopSpinner("Codebase scanned successfully")
	return result, nil
}

// ShowScanResults displays scan results to the user
func (s *service) ShowScanResults(result *scanner.ScanResult) error {
	// Display results
	stats := s.scanner.GetStatistics(result)
	fmt.Printf("\nScan Results:\n")
	fmt.Printf("  • Handlers found: %d\n", stats.HandlersFound)
	fmt.Printf("  • Routes found: %d\n", stats.RoutesFound)
	fmt.Printf("  • Providers found: %d\n", stats.ProvidersFound)
	fmt.Printf("  • Packages scanned: %d\n", stats.PackagesScanned)

	if stats.ErrorsFound > 0 {
		fmt.Printf("  • Errors: %d\n", stats.ErrorsFound)
	}

	// Show detailed results if requested
	if len(result.Handlers) > 0 {
		fmt.Println("\nHandlers:")
		for _, h := range result.Handlers {
			fmt.Printf("  - %s.%s (%s)\n", h.Package, h.FunctionName, h.HandlerName)
		}
	}

	if len(result.Routes) > 0 {
		fmt.Println("\nRoutes:")
		for _, r := range result.Routes {
			// Convert {param} to :param for display consistency with generated routes
			displayPath := strings.ReplaceAll(r.Path, "{", ":")
			displayPath = strings.ReplaceAll(displayPath, "}", "")
			fmt.Printf("  - %s %s -> %s\n", r.HTTPMethod, displayPath, r.HandlerRef)
		}
	}

	if len(result.Providers) > 0 {
		fmt.Println("\nProviders:")
		for _, p := range result.Providers {
			fmt.Printf("  - %s() -> %s\n", p.FunctionName, p.ReturnType)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range result.Errors {
			fmt.Printf("  - %s: %s\n", e.FilePath, e.Message)
		}
	}

	return nil
}

// ValidateScanResults performs validation on scan results
func (s *service) ValidateScanResults(result *scanner.ScanResult) error {
	validator := scanner.NewValidator()
	validation := validator.ValidateScanResult(result)

	if validation.HasErrors() {
		fmt.Println("\nValidation Errors:")
		for _, err := range validation.Errors {
			fmt.Printf("  • %s: %s\n", err.Type, err.Message)
		}
	}

	if validation.HasWarnings() {
		fmt.Println("\nValidation Warnings:")
		for _, warn := range validation.Warnings {
			fmt.Printf("  • %s: %s\n", warn.Type, warn.Message)
		}
	}

	return nil
}
