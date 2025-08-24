package generator

import (
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
		routesByPackage[route.Package] = append(routesByPackage[route.Package], route)
	}

	// Sort routes within each package by path for consistent output
	for pkg := range routesByPackage {
		sort.Slice(routesByPackage[pkg], func(i, j int) bool {
			return routesByPackage[pkg][i].Path < routesByPackage[pkg][j].Path
		})
	}

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
	var allRoutes []scanner.RouteMapping
	for _, routes := range routesByPackage {
		allRoutes = append(allRoutes, routes...)
	}

	// Sort routes by path for consistent output
	sort.Slice(allRoutes, func(i, j int) bool {
		if allRoutes[i].Path == allRoutes[j].Path {
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

	tmpl, err := template.New("routes").Parse(routeTemplate)
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
func (g *RouteGenerator) getAPIPrefix(path string) string {
	// Extract prefix like /api/v1 from paths
	parts := strings.Split(path, "/")
	if len(parts) >= 3 && parts[1] == "api" {
		return "/" + parts[1] + "/" + parts[2] // /api/v1
	}
	return "/api/v1" // Default fallback
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

	fmt.Printf("✅ Generated: %s\n", path)
	return nil
}

// routeTemplate is the Go template for generating the routes file
const routeTemplate = `// Code generated by taskw. DO NOT EDIT.

package {{.Package}}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)

// RegisterRoutes registers all HTTP routes with the Fiber app
func (s *Server) RegisterRoutes(app *fiber.App) {
	{{- range $routes := .Routes}}
	app.{{call $.GetRouterMethod .HTTPMethod}}("{{.Path}}", {{call $.GetHandlerRef .Package .HandlerRef}})
	{{- end}}

	s.logger.Info("All routes registered successfully")
}
`
