package generator

import (
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/nkaewam/taskw/internal/config"
	"github.com/nkaewam/taskw/internal/scanner"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// RouteGenerator generates Fiber route registration code
type RouteGenerator struct {
	config *config.Config
}

// NewRouteGenerator creates a new route generator
func NewRouteGenerator(cfg *config.Config) *RouteGenerator {
	return &RouteGenerator{
		config: cfg,
	}
}

// GenerateRoutes generates the routes_gen.go file
func (g *RouteGenerator) GenerateRoutes(handlers []scanner.HandlerFunction, routes []scanner.RouteMapping) error {
	if !g.config.Generation.Routes.Enabled {
		return nil
	}

	// Organize routes by package for better structure
	routesByPackage := g.organizeRoutesByPackage(routes)

	// Generate imports needed
	imports := g.generateImports(handlers, routes)

	// Get output path
	outputPath := filepath.Join(g.config.Paths.OutputDir, g.config.Generation.Routes.OutputFile)

	// Generate the file content
	content, err := g.generateRouteFileContent(routesByPackage, imports)
	if err != nil {
		return fmt.Errorf("error generating route file content: %w", err)
	}

	// Write to file (assuming a file writer utility will be available)
	return writeGeneratedFile(outputPath, content)
}

// organizeRoutesByPackage groups routes by their package for better organization
func (g *RouteGenerator) organizeRoutesByPackage(routes []scanner.RouteMapping) map[string][]scanner.RouteMapping {
	routesByPackage := make(map[string][]scanner.RouteMapping)

	for _, route := range routes {
		// Convert path format early for consistent sorting
		route.Path = g.convertPathForFiber(route.Path)
		routesByPackage[route.Package] = append(routesByPackage[route.Package], route)
	}

	// Routes will be sorted globally later

	return routesByPackage
}

// generateImports creates the import statements needed for the generated file
func (g *RouteGenerator) generateImports(handlers []scanner.HandlerFunction, routes []scanner.RouteMapping) []string {
	imports := []string{
		`"github.com/gofiber/fiber/v2"`,
	}

	// For route generation, we only need fiber import
	// Handler references like s.userHandler.GetUsers don't require importing the handler packages
	// since they're just struct field references

	return imports
}

// generateRouteFileContent creates the actual file content
func (g *RouteGenerator) generateRouteFileContent(routesByPackage map[string][]scanner.RouteMapping, imports []string) (string, error) {
	// Flatten routes from all packages into a single slice
	// Process packages in deterministic order
	var packageNames []string
	for pkg := range routesByPackage {
		packageNames = append(packageNames, pkg)
	}
	sort.Strings(packageNames)

	var allRoutes []scanner.RouteMapping
	for _, pkg := range packageNames {
		allRoutes = append(allRoutes, routesByPackage[pkg]...)
	}

	// Sort routes with more specific routes first to avoid conflicts
	// This is the final sort that determines the order in the generated file
	sort.Slice(allRoutes, func(i, j int) bool {
		scoreA := g.calculateSpecificityScore(allRoutes[i].Path)
		scoreB := g.calculateSpecificityScore(allRoutes[j].Path)

		// Higher score means more specific (should come first)
		if scoreA != scoreB {
			return scoreA > scoreB
		}

		// If scores are equal, sort by HTTP method then path
		if allRoutes[i].HTTPMethod != allRoutes[j].HTTPMethod {
			return allRoutes[i].HTTPMethod < allRoutes[j].HTTPMethod
		}

		return allRoutes[i].Path < allRoutes[j].Path
	})

	data := struct {
		Package         string
		Imports         []string
		Routes          []scanner.RouteMapping
		GetRouterMethod func(method string) string
		GetHandlerRef   func(pkg, handlerRef string) string
	}{
		Package:         "api",
		Imports:         imports,
		Routes:          allRoutes,
		GetRouterMethod: g.getRouterMethod,
		GetHandlerRef:   g.getHandlerRef,
	}

	tmplContent, err := templateFS.ReadFile("templates/routes.tmpl")
	if err != nil {
		return "", fmt.Errorf("error reading route template: %w", err)
	}

	tmpl, err := template.New("routes").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("error parsing route template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing route template: %w", err)
	}

	return buf.String(), nil
}

// organizeRoutesByAPIGroups groups routes by their API prefix
// Unused for now, but can be used in the future
func (g *RouteGenerator) organizeRoutesByAPIGroups(routesByPackage map[string][]scanner.RouteMapping) map[string][]scanner.RouteMapping {
	apiGroups := make(map[string][]scanner.RouteMapping)

	for _, routes := range routesByPackage {
		for _, route := range routes {
			prefix := g.getAPIPrefix(route.Path)
			apiGroups[prefix] = append(apiGroups[prefix], route)
		}
	}

	// Sort routes within each group
	for prefix := range apiGroups {
		sort.Slice(apiGroups[prefix], func(i, j int) bool {
			if apiGroups[prefix][i].Path == apiGroups[prefix][j].Path {
				return apiGroups[prefix][i].HTTPMethod < apiGroups[prefix][j].HTTPMethod
			}
			return apiGroups[prefix][i].Path < apiGroups[prefix][j].Path
		})
	}

	return apiGroups
}

// getRelativePath extracts the relative path after removing the prefix
// Unused for now, but can be used in the future
func (g *RouteGenerator) getRelativePath(fullPath, prefix string) string {
	if strings.HasPrefix(fullPath, prefix) {
		relativePath := fullPath[len(prefix):]
		if relativePath == "" {
			return "/"
		}
		// Convert {param} to :param for Fiber
		relativePath = strings.ReplaceAll(relativePath, "{", ":")
		relativePath = strings.ReplaceAll(relativePath, "}", "")
		return relativePath
	}
	return fullPath
}

// getAPIPrefix extracts the API prefix from a route path
// Unused for now, but can be used in the future
func (g *RouteGenerator) getAPIPrefix(path string) string {
	// Extract prefix like /api/v1 from paths
	parts := strings.Split(path, "/")
	if len(parts) >= 3 && parts[1] == "api" {
		return "/" + parts[1] + "/" + parts[2] // /api/v1
	}
	return "/" // Default fallback
}

// getRouterMethod maps HTTP methods to Fiber router methods
func (g *RouteGenerator) getRouterMethod(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return "Get"
	case "POST":
		return "Post"
	case "PUT":
		return "Put"
	case "DELETE":
		return "Delete"
	case "PATCH":
		return "Patch"
	case "HEAD":
		return "Head"
	case "OPTIONS":
		return "Options"
	default:
		return "All" // Fallback for unsupported methods
	}
}

// getHandlerRef generates the handler reference for route registration
func (g *RouteGenerator) getHandlerRef(pkg, handlerRef string) string {
	// handlerRef comes from scanner as "userHandler.GetUsers"
	// We need to convert it to "s.userHandler.GetUsers"
	parts := strings.Split(handlerRef, ".")
	if len(parts) == 2 {
		handlerName := parts[0] // e.g., "userHandler"
		methodName := parts[1]  // e.g., "GetUsers"
		return fmt.Sprintf("s.%s.%s", handlerName, methodName)
	}
	return handlerRef
}

// convertPathForFiber converts OpenAPI/Swagger path parameters to Fiber format
// Converts {param} to :param for Fiber router
func (g *RouteGenerator) convertPathForFiber(path string) string {
	// Convert {param} to :param for Fiber
	converted := strings.ReplaceAll(path, "{", ":")
	converted = strings.ReplaceAll(converted, "}", "")
	return converted
}

// isMoreSpecificRoute determines if pathA is more specific than pathB
// More specific routes should be registered first to avoid conflicts
func (g *RouteGenerator) isMoreSpecificRoute(pathA, pathB string) bool {
	// Calculate specificity scores for both paths
	scoreA := g.calculateSpecificityScore(pathA)
	scoreB := g.calculateSpecificityScore(pathB)

	// Higher score means more specific
	if scoreA != scoreB {
		return scoreA > scoreB
	}

	// If scores are equal, use alphabetical order for consistency
	return pathA < pathB
}

// calculateSpecificityScore calculates a numeric score for route specificity
// Higher scores indicate more specific routes that should be registered first
func (g *RouteGenerator) calculateSpecificityScore(path string) int {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	score := 0

	// Base score: longer paths are more specific
	score += len(segments) * 1000

	// Bonus for static segments, penalty for parameters
	for _, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			score -= 100 // Parameter penalty
		} else {
			score += 100 // Static segment bonus
		}
	}

	return score
}

// countPathParameters counts the number of parameters in a path
func (g *RouteGenerator) countPathParameters(path string) int {
	count := 0
	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			count++
		}
	}
	return count
}

// writeGeneratedFile writes content to a file with proper Go formatting
func writeGeneratedFile(path, content string) error {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Format the generated Go code
	formatted, err := format.Source([]byte(content))
	if err != nil {
		// If formatting fails, still write the unformatted content
		// This helps with debugging template issues
		fmt.Printf("Warning: Failed to format generated code: %v\n", err)
		formatted = []byte(content)
	}

	// Write the file
	if err := os.WriteFile(path, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}
