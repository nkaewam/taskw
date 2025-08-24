package generator

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/nkaewam/taskw/internal/config"
	"github.com/nkaewam/taskw/internal/scanner"
)

// DependencyGenerator generates Wire provider sets
type DependencyGenerator struct {
	config *config.Config
}

// NewDependencyGenerator creates a new dependency generator
func NewDependencyGenerator(cfg *config.Config) *DependencyGenerator {
	return &DependencyGenerator{
		config: cfg,
	}
}

// GenerateDependencies generates the dependencies_gen.go file
func (g *DependencyGenerator) GenerateDependencies(providers []scanner.ProviderFunction) error {
	if !g.config.Generation.Dependencies.Enabled {
		return nil
	}

	// Organize providers by package for better structure
	providersByPackage := g.organizeProvidersByPackage(providers)

	// Generate imports needed
	imports := g.generateImports(providers)

	// Get output path
	outputPath := filepath.Join(g.config.Paths.OutputDir, g.config.Generation.Dependencies.OutputFile)

	// Generate the file content
	content, err := g.generateDependencyFileContent(providersByPackage, imports)
	if err != nil {
		return fmt.Errorf("error generating dependency file content: %w", err)
	}

	// Write to file
	return writeGeneratedFile(outputPath, content)
}

// organizeProvidersByPackage groups providers by their package
func (g *DependencyGenerator) organizeProvidersByPackage(providers []scanner.ProviderFunction) map[string][]scanner.ProviderFunction {
	providersByPackage := make(map[string][]scanner.ProviderFunction)

	for _, provider := range providers {
		providersByPackage[provider.Package] = append(providersByPackage[provider.Package], provider)
	}

	// Sort providers within each package by function name for consistent output
	for pkg := range providersByPackage {
		sort.Slice(providersByPackage[pkg], func(i, j int) bool {
			return providersByPackage[pkg][i].FunctionName < providersByPackage[pkg][j].FunctionName
		})
	}

	return providersByPackage
}

// generateImports creates the import statements needed for the generated file
func (g *DependencyGenerator) generateImports(providers []scanner.ProviderFunction) []string {
	imports := []string{
		`"github.com/google/wire"`,
	}

	// Determine the output package name from the output directory
	outputPackage := g.getOutputPackageName()

	// Collect unique packages that need to be imported
	packageSet := make(map[string]bool)
	for _, provider := range providers {
		if provider.Package != "" && provider.Package != outputPackage {
			// Generate import path based on module and package
			// Skip importing the same package as the output file
			importPath := fmt.Sprintf(`"%s/internal/%s"`, g.config.Project.Module, provider.Package)
			packageSet[importPath] = true
		}
	}

	// Convert to sorted slice
	for pkg := range packageSet {
		imports = append(imports, pkg)
	}

	sort.Strings(imports[1:]) // Sort everything except wire import
	return imports
}

// generateDependencyFileContent creates the actual file content
func (g *DependencyGenerator) generateDependencyFileContent(providersByPackage map[string][]scanner.ProviderFunction, imports []string) (string, error) {
	data := struct {
		Package            string
		Imports            []string
		ProvidersByPackage map[string][]scanner.ProviderFunction
		GetProviderRef     func(pkg, functionName string) string
	}{
		Package:            "api",
		Imports:            imports,
		ProvidersByPackage: providersByPackage,
		GetProviderRef:     g.getProviderRef,
	}

	tmplContent, err := templateFS.ReadFile("templates/dependencies.tmpl")
	if err != nil {
		return "", fmt.Errorf("error reading dependency template: %w", err)
	}

	tmpl, err := template.New("dependencies").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("error parsing dependency template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing dependency template: %w", err)
	}

	return buf.String(), nil
}

// getProviderRef generates the provider reference for Wire
func (g *DependencyGenerator) getProviderRef(pkg, functionName string) string {
	outputPackage := g.getOutputPackageName()

	// If the provider is in the same package as the output file,
	// don't use the package prefix
	if pkg == outputPackage {
		return functionName
	}

	return fmt.Sprintf("%s.%s", pkg, functionName)
}

// getOutputPackageName determines the package name of the output file
func (g *DependencyGenerator) getOutputPackageName() string {
	// Extract package name from output directory
	// e.g., "./internal/api" -> "api"
	outputDir := g.config.Paths.OutputDir
	return filepath.Base(outputDir)
}
