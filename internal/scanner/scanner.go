package scanner

import (
	"fmt"
	"sync"

	"github.com/nkaewam/taskw/internal/config"
)

// Scanner is the main hybrid scanner that combines file filtering with AST parsing
type Scanner struct {
	config     *config.Config
	astScanner *ASTScanner
	fileFilter *FileFilter
}

// NewScanner creates a new hybrid scanner instance
func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{
		config:     cfg,
		astScanner: NewASTScanner(cfg.Paths.ScanDirs),
		fileFilter: NewFileFilter(),
	}
}

// ScanAll scans all configured directories for handlers, routes, and providers
func (s *Scanner) ScanAll() (*ScanResult, error) {
	result := &ScanResult{
		Handlers:  []HandlerFunction{},
		Routes:    []RouteMapping{},
		Providers: []ProviderFunction{},
		Errors:    []ScanError{},
	}

	// Scan all configured directories
	for _, dir := range s.config.Paths.ScanDirs {
		dirResult, err := s.ScanDirectory(dir)
		if err != nil {
			return nil, fmt.Errorf("error scanning directory %s: %w", dir, err)
		}

		// Merge results
		result.Handlers = append(result.Handlers, dirResult.Handlers...)
		result.Routes = append(result.Routes, dirResult.Routes...)
		result.Providers = append(result.Providers, dirResult.Providers...)
		result.Errors = append(result.Errors, dirResult.Errors...)
	}

	return result, nil
}

// ScanDirectory scans a single directory using the hybrid approach
func (s *Scanner) ScanDirectory(directory string) (*ScanResult, error) {
	// Step 1: Use file filter to find candidate files
	candidateFiles, err := s.fileFilter.FindCandidateFiles(directory)
	if err != nil {
		return nil, fmt.Errorf("error finding candidate files in %s: %w", directory, err)
	}

	// Step 2: Parse candidate files with AST scanner (parallel processing)
	return s.scanFilesParallel(candidateFiles), nil
}

// ScanRoutes specifically scans for handlers and routes (for backwards compatibility)
func (s *Scanner) ScanRoutes(directories []string) ([]HandlerFunction, []RouteMapping, error) {
	var allHandlers []HandlerFunction
	var allRoutes []RouteMapping

	for _, dir := range directories {
		result, err := s.ScanDirectory(dir)
		if err != nil {
			return nil, nil, err
		}

		allHandlers = append(allHandlers, result.Handlers...)
		allRoutes = append(allRoutes, result.Routes...)
	}

	return allHandlers, allRoutes, nil
}

// ScanProviders specifically scans for provider functions
func (s *Scanner) ScanProviders(directories []string) ([]ProviderFunction, error) {
	var allProviders []ProviderFunction

	for _, dir := range directories {
		result, err := s.ScanDirectory(dir)
		if err != nil {
			return nil, err
		}

		allProviders = append(allProviders, result.Providers...)
	}

	return allProviders, nil
}

// scanFilesParallel processes multiple files in parallel for better performance
func (s *Scanner) scanFilesParallel(files []string) *ScanResult {
	result := &ScanResult{
		Handlers:  []HandlerFunction{},
		Routes:    []RouteMapping{},
		Providers: []ProviderFunction{},
		Errors:    []ScanError{},
	}

	// Use a reasonable number of goroutines to avoid overwhelming the system
	const maxGoroutines = 10
	sem := make(chan struct{}, maxGoroutines)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, file := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Scan the file
			fileResult, err := s.astScanner.ScanFile(filePath)
			if err != nil {
				// Add error to results but continue processing
				mu.Lock()
				result.Errors = append(result.Errors, ScanError{
					FilePath: filePath,
					Message:  err.Error(),
					Type:     "parse_error",
				})
				mu.Unlock()
				return
			}

			// Merge results thread-safely
			mu.Lock()
			result.Handlers = append(result.Handlers, fileResult.Handlers...)
			result.Routes = append(result.Routes, fileResult.Routes...)
			result.Providers = append(result.Providers, fileResult.Providers...)
			result.Errors = append(result.Errors, fileResult.Errors...)
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	return result
}

// GetStatistics returns scanning statistics for debugging
func (s *Scanner) GetStatistics(result *ScanResult) ScanStatistics {
	return ScanStatistics{
		HandlersFound:   len(result.Handlers),
		RoutesFound:     len(result.Routes),
		ProvidersFound:  len(result.Providers),
		ErrorsFound:     len(result.Errors),
		PackagesScanned: s.countUniquePackages(result),
	}
}

// ScanStatistics provides information about scanning results
type ScanStatistics struct {
	HandlersFound   int
	RoutesFound     int
	ProvidersFound  int
	ErrorsFound     int
	PackagesScanned int
}

func (s *Scanner) countUniquePackages(result *ScanResult) int {
	packages := make(map[string]bool)

	for _, h := range result.Handlers {
		packages[h.Package] = true
	}

	for _, p := range result.Providers {
		packages[p.Package] = true
	}

	return len(packages)
}
