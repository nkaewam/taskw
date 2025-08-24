package scanner

import (
	"fmt"
	"strings"
)

// ValidationResult contains validation errors and warnings
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationWarning
}

// ValidationError represents a validation error that prevents code generation
type ValidationError struct {
	Type     string // "duplicate_route", "invalid_signature", etc.
	Message  string
	FilePath string
	Line     int
	Handler  *HandlerFunction
	Route    *RouteMapping
}

// ValidationWarning represents a validation warning that might cause issues
type ValidationWarning struct {
	Type     string
	Message  string
	FilePath string
	Handler  *HandlerFunction
}

// Validator validates scan results for common issues
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateScanResult validates handlers, routes, and providers for common issues
func (v *Validator) ValidateScanResult(result *ScanResult) *ValidationResult {
	validationResult := &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Validate routes
	v.validateRoutes(result.Routes, validationResult)

	// Validate handlers
	v.validateHandlers(result.Handlers, validationResult)

	// Validate handler-route matching
	v.validateHandlerRouteMatching(result.Handlers, result.Routes, validationResult)

	return validationResult
}

// validateRoutes checks for duplicate routes and invalid route patterns
func (v *Validator) validateRoutes(routes []RouteMapping, result *ValidationResult) {
	routeMap := make(map[string][]RouteMapping)

	for _, route := range routes {
		key := fmt.Sprintf("%s %s", route.HTTPMethod, route.Path)
		routeMap[key] = append(routeMap[key], route)
	}

	// Check for duplicate routes
	for key, duplicates := range routeMap {
		if len(duplicates) > 1 {
			for _, dup := range duplicates {
				result.Errors = append(result.Errors, ValidationError{
					Type:    "duplicate_route",
					Message: fmt.Sprintf("Duplicate route found: %s", key),
					Route:   &dup,
				})
			}
		}
	}

	// Validate route patterns
	for _, route := range routes {
		if err := v.validateRoutePattern(route); err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "invalid_route_pattern",
				Message: err.Error(),
				Route:   &route,
			})
		}
	}
}

// validateHandlers checks handler function signatures and naming conventions
func (v *Validator) validateHandlers(handlers []HandlerFunction, result *ValidationResult) {
	for _, handler := range handlers {
		// Check naming conventions
		if !strings.HasSuffix(handler.HandlerName, "Handler") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:     "naming_convention",
				Message:  fmt.Sprintf("Handler struct %s should end with 'Handler'", handler.HandlerName),
				FilePath: handler.FilePath,
				Handler:  &handler,
			})
		}

		// Check for common patterns in function names
		if strings.Contains(strings.ToLower(handler.FunctionName), "test") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:     "test_function",
				Message:  fmt.Sprintf("Function %s appears to be a test function but was detected as a handler", handler.FunctionName),
				FilePath: handler.FilePath,
				Handler:  &handler,
			})
		}
	}
}

// validateHandlerRouteMatching ensures handlers have corresponding routes and vice versa
func (v *Validator) validateHandlerRouteMatching(handlers []HandlerFunction, routes []RouteMapping, result *ValidationResult) {
	handlerMap := make(map[string]HandlerFunction)
	routeMap := make(map[string]RouteMapping)

	// Build lookup maps
	for _, handler := range handlers {
		key := fmt.Sprintf("%s.%s", handler.Package, handler.FunctionName)
		handlerMap[key] = handler
	}

	for _, route := range routes {
		key := fmt.Sprintf("%s.%s", route.Package, route.MethodName)
		routeMap[key] = route
	}

	// Check for handlers without routes
	for key, handler := range handlerMap {
		if _, exists := routeMap[key]; !exists {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:     "handler_without_route",
				Message:  fmt.Sprintf("Handler function %s.%s found but no @Router annotation", handler.Package, handler.FunctionName),
				FilePath: handler.FilePath,
				Handler:  &handler,
			})
		}
	}

	// Check for routes without handlers
	for key, route := range routeMap {
		if _, exists := handlerMap[key]; !exists {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "route_without_handler",
				Message: fmt.Sprintf("@Router annotation found for %s.%s but no corresponding handler function", route.Package, route.MethodName),
				Route:   &route,
			})
		}
	}
}

// validateRoutePattern validates Fiber route patterns
func (v *Validator) validateRoutePattern(route RouteMapping) error {
	path := route.Path

	// Basic validation
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("route path must start with '/': %s", path)
	}

	// Check for invalid characters in static parts
	// Fiber supports :param and *wildcard, but let's do basic validation
	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if segment == "" {
			continue
		}

		// Skip parameter and wildcard segments
		if strings.HasPrefix(segment, ":") || strings.HasPrefix(segment, "*") {
			continue
		}

		// Check for invalid characters in static segments
		if strings.ContainsAny(segment, " \t\n\r") {
			return fmt.Errorf("route path contains whitespace: %s", path)
		}
	}

	// Validate HTTP method
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	methodValid := false
	for _, validMethod := range validMethods {
		if route.HTTPMethod == validMethod {
			methodValid = true
			break
		}
	}

	if !methodValid {
		return fmt.Errorf("invalid HTTP method: %s", route.HTTPMethod)
	}

	return nil
}

// HasErrors returns true if there are validation errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// HasWarnings returns true if there are validation warnings
func (vr *ValidationResult) HasWarnings() bool {
	return len(vr.Warnings) > 0
}
