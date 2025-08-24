package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// FileFilter handles filtering of Go files based on .taskwignore patterns
type FileFilter struct {
	ignorePatterns []string
	defaultIgnores []string
}

// NewFileFilter creates a new file filter and loads .taskwignore patterns
func NewFileFilter() *FileFilter {
	filter := &FileFilter{
		defaultIgnores: []string{
			"vendor/**",
			"node_modules/**",
			".git/**",
			".task/**",
			"bin/**",
			"build/**",
			"dist/**",
			"**/*_test.go",   // Exclude test files
			"**/*_mock.go",   // Exclude mock files
			"**/testdata/**", // Exclude test data
		},
	}

	// Load .taskwignore patterns
	filter.loadTaskwIgnore()
	return filter
}

// loadTaskwIgnore reads .taskwignore file and loads ignore patterns
func (f *FileFilter) loadTaskwIgnore() {
	f.ignorePatterns = make([]string, len(f.defaultIgnores))
	copy(f.ignorePatterns, f.defaultIgnores)

	file, err := os.Open(".taskwignore")
	if err != nil {
		// .taskwignore doesn't exist, use only default patterns
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		f.ignorePatterns = append(f.ignorePatterns, line)
	}
}

// FindCandidateFiles recursively finds all Go files that are not ignored
func (f *FileFilter) FindCandidateFiles(rootDir string) ([]string, error) {
	var candidates []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from root directory for pattern matching
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		// Skip directories that match ignore patterns
		if info.IsDir() {
			if f.shouldIgnore(relPath) {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Check if file should be ignored
		if !f.shouldIgnore(relPath) {
			candidates = append(candidates, path)
		}

		return nil
	})

	return candidates, err
}

// shouldIgnore checks if a file or directory path matches any ignore pattern
func (f *FileFilter) shouldIgnore(relPath string) bool {
	// Normalize path separators to forward slashes for consistent matching
	normalizedPath := filepath.ToSlash(relPath)

	for _, pattern := range f.ignorePatterns {
		if f.matchPattern(pattern, normalizedPath) {
			return true
		}
	}

	return false
}

// matchPattern implements basic glob pattern matching for ignore patterns
func (f *FileFilter) matchPattern(pattern, path string) bool {
	// Handle ** patterns (match any number of directories)
	if strings.Contains(pattern, "**") {
		return f.matchDoubleStarPattern(pattern, path)
	}

	// Handle * patterns (match within a single directory level)
	if strings.Contains(pattern, "*") {
		return f.matchSingleStarPattern(pattern, path)
	}

	// Exact match or directory prefix match
	if pattern == path {
		return true
	}

	// Check if pattern matches as directory prefix
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(path, pattern) || strings.HasPrefix(path+"/", pattern)
	}

	// Check if path starts with pattern as directory
	return strings.HasPrefix(path, pattern+"/")
}

// matchDoubleStarPattern handles ** glob patterns
func (f *FileFilter) matchDoubleStarPattern(pattern, path string) bool {
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		// Complex ** patterns not supported yet, fall back to simple matching
		return strings.Contains(path, strings.ReplaceAll(pattern, "**", ""))
	}

	prefix := strings.TrimSuffix(parts[0], "/")
	suffix := strings.TrimPrefix(parts[1], "/")

	// Check prefix match
	if prefix != "" && !strings.HasPrefix(path, prefix) {
		return false
	}

	// Check suffix match
	if suffix != "" {
		if strings.Contains(suffix, "*") {
			// Suffix has more wildcards, handle recursively
			pathWithoutPrefix := path
			if prefix != "" {
				pathWithoutPrefix = strings.TrimPrefix(path, prefix+"/")
			}
			return f.matchSingleStarPattern(suffix, pathWithoutPrefix)
		}
		return strings.HasSuffix(path, suffix)
	}

	return true
}

// matchSingleStarPattern handles * glob patterns (single directory level)
func (f *FileFilter) matchSingleStarPattern(pattern, path string) bool {
	// Split pattern and path by slashes for directory-level matching
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	return f.matchParts(patternParts, pathParts)
}

// matchParts matches pattern parts against path parts
func (f *FileFilter) matchParts(patternParts, pathParts []string) bool {
	if len(patternParts) > len(pathParts) {
		return false
	}

	for i, patternPart := range patternParts {
		if i >= len(pathParts) {
			return false
		}

		if !f.matchSinglePart(patternPart, pathParts[i]) {
			return false
		}
	}

	return true
}

// matchSinglePart matches a single pattern part with wildcards against a path part
func (f *FileFilter) matchSinglePart(pattern, pathPart string) bool {
	if pattern == "*" {
		return true
	}

	if !strings.Contains(pattern, "*") {
		return pattern == pathPart
	}

	// Handle simple * patterns within a single part
	return f.wildcardMatch(pattern, pathPart)
}

// wildcardMatch implements simple wildcard matching for a single pattern part
func (f *FileFilter) wildcardMatch(pattern, str string) bool {
	// Simple implementation for basic * wildcards
	parts := strings.Split(pattern, "*")

	index := 0
	for i, part := range parts {
		if part == "" {
			continue
		}

		pos := strings.Index(str[index:], part)
		if pos == -1 {
			return false
		}

		// For first part, must match at beginning
		if i == 0 && pos != 0 {
			return false
		}

		index += pos + len(part)
	}

	// For last part, must match at end
	if len(parts) > 0 && parts[len(parts)-1] != "" {
		return strings.HasSuffix(str, parts[len(parts)-1])
	}

	return true
}

// CreateDefaultTaskwIgnore creates a default .taskwignore file
func (f *FileFilter) CreateDefaultTaskwIgnore() error {
	content := `# TaskW Ignore Patterns
# This file specifies which files and directories to ignore when scanning for handlers and providers
# Patterns follow gitignore-style glob syntax

# Dependencies and vendor code
vendor/**
node_modules/**

# Build artifacts
bin/**
build/**
dist/**
*.exe
*.dll
*.so
*.dylib

# Test files and test data
**/*_test.go
**/*_mock.go
**/testdata/**
**/mocks/**

# IDE and tool files
.git/**
.vscode/**
.idea/**
*.swp
*.swo
*~

# Logs and temporary files
*.log
*.tmp
tmp/**

# Generated files (add your specific patterns)
# **/*_gen.go
# **/generated/**
`

	return os.WriteFile(".taskwignore", []byte(content), 0644)
}
