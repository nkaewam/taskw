package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

// ASTScanner uses Go's AST parser for accurate code analysis
type ASTScanner struct {
	fset *token.FileSet
}

// NewASTScanner creates a new AST-based scanner
func NewASTScanner() *ASTScanner {
	return &ASTScanner{
		fset: token.NewFileSet(),
	}
}

// ScanFile parses a Go file and extracts handlers, routes, and providers
func (s *ASTScanner) ScanFile(filePath string) (*ScanResult, error) {
	// Parse the Go file into AST
	node, err := parser.ParseFile(s.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	result := &ScanResult{
		Handlers:        []HandlerFunction{},
		Routes:          []RouteMapping{},
		Providers:       []ProviderFunction{},
		Interfaces:      []HandlerInterface{},
		Implementations: []HandlerImplementation{},
		Errors:          []ScanError{},
	}

	packageName := node.Name.Name

	// Walk the AST to find functions and type declarations
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			s.processFuncDecl(x, packageName, filePath, result)
		case *ast.TypeSpec:
			s.processTypeSpec(x, packageName, filePath, result)
		}
		return true
	})

	// After scanning all types and functions, associate interfaces with implementations
	s.associateInterfacesWithImplementations(result)

	return result, nil
}

// processFuncDecl analyzes a function declaration for handlers and providers
func (s *ASTScanner) processFuncDecl(fn *ast.FuncDecl, pkg, filePath string, result *ScanResult) {
	// Check if this is a handler function
	if handler := s.extractHandler(fn, pkg, filePath); handler != nil {
		result.Handlers = append(result.Handlers, *handler)

		// Look for @Router annotation
		if route := s.extractRoute(fn, *handler); route != nil {
			result.Routes = append(result.Routes, *route)
		}
	}

	// Check if this is a provider function
	if provider := s.extractProvider(fn, pkg, filePath); provider != nil {
		result.Providers = append(result.Providers, *provider)
	}
}

// extractHandler checks if a function is a Fiber handler and extracts its information
func (s *ASTScanner) extractHandler(fn *ast.FuncDecl, pkg, filePath string) *HandlerFunction {
	// Must have a receiver
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return nil
	}

	// Check receiver type: should be *SomeHandler or *SomeImpl (for interface pattern)
	recv := fn.Recv.List[0]
	handlerName := s.getReceiverTypeName(recv)
	if handlerName == "" {
		return nil
	}

	// Accept both traditional pattern (*Handler) and interface pattern (*Impl)
	if !strings.HasSuffix(handlerName, "Handler") && !s.isHandlerImplementation(handlerName) {
		return nil
	}

	// Check function parameters: should have (c *fiber.Ctx)
	if !s.hasFiberCtxParam(fn) {
		return nil
	}

	// Check return type: should return error
	if !s.returnsError(fn) {
		return nil
	}

	return &HandlerFunction{
		FunctionName: fn.Name.Name,
		Package:      pkg,
		HandlerName:  handlerName,
		ReturnType:   "error",
		FilePath:     filePath,
	}
}

// extractRoute parses @Router comments to extract route information
// Supports multiple standard Swagger annotation formats:
// - @Router /path [method]
// - @Router "/path" [method]
// - @router /path [method] (case insensitive)
func (s *ASTScanner) extractRoute(fn *ast.FuncDecl, handler HandlerFunction) *RouteMapping {
	if fn.Doc == nil {
		return nil
	}

	// Enhanced regex patterns for standard Swagger formats
	routerPatterns := []*regexp.Regexp{
		// Standard format: @Router /path [method]
		regexp.MustCompile(`(?i)@Router\s+([^\s\[\]]+)\s+\[([^\]]+)\]`),
		// Quoted path format: @Router "/path" [method]
		regexp.MustCompile(`(?i)@Router\s+"([^"]+)"\s+\[([^\]]+)\]`),
		// Alternative format: @Router /path method
		regexp.MustCompile(`(?i)@Router\s+([^\s]+)\s+([A-Za-z]+)(?:\s|$)`),
		// Gin-style format: @router /path [method]
		regexp.MustCompile(`(?i)@router\s+([^\s\[\]]+)\s+\[([^\]]+)\]`),
	}

	for _, comment := range fn.Doc.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		text = strings.TrimSpace(strings.TrimPrefix(text, "*")) // Support /** comments

		for _, pattern := range routerPatterns {
			if matches := pattern.FindStringSubmatch(text); matches != nil {
				path := strings.Trim(matches[1], `"'`) // Remove quotes if present
				method := strings.ToUpper(strings.TrimSpace(matches[2]))

				// Validate HTTP method
				if !s.isValidHTTPMethod(method) {
					continue
				}

				return &RouteMapping{
					MethodName: fn.Name.Name,
					Path:       path,
					HTTPMethod: method,
					HandlerRef: s.generateHandlerRef(handler),
					Package:    handler.Package,
				}
			}
		}
	}

	return nil
}

// generateHandlerRef creates a proper handler reference
func (s *ASTScanner) generateHandlerRef(handler HandlerFunction) string {
	// Use package name as the base for handler reference
	// e.g., "user" package becomes "userHandler"
	handlerName := handler.Package + "Handler"

	// Convert first letter to lowercase for field reference
	if len(handlerName) > 0 {
		handlerName = strings.ToLower(handlerName[:1]) + handlerName[1:]
	}

	return fmt.Sprintf("%s.%s", handlerName, handler.FunctionName)
}

// isValidHTTPMethod checks if the method is a valid HTTP method
func (s *ASTScanner) isValidHTTPMethod(method string) bool {
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
		"TRACE":   true,
		"CONNECT": true,
	}
	return validMethods[method]
}

// extractProvider checks if a function is a Wire provider function
func (s *ASTScanner) extractProvider(fn *ast.FuncDecl, pkg, filePath string) *ProviderFunction {
	// Must start with "Provide"
	if !strings.HasPrefix(fn.Name.Name, "Provide") {
		return nil
	}

	// Must have return type(s)
	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return nil
	}

	// Extract return type (first return value)
	returnType := s.getTypeString(fn.Type.Results.List[0].Type)
	if returnType == "" {
		return nil
	}

	// For interface-based patterns, if the return type is just "Handler" (interface),
	// we should qualify it with the package name for clarity in generated code
	if returnType == "Handler" && s.hasErrorReturnType(fn) {
		// This looks like it returns (Handler, error) - an interface pattern
		returnType = pkg + "." + returnType
	}

	// Extract parameters
	var parameters []string
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			paramType := s.getTypeString(param.Type)
			if paramType != "" {
				parameters = append(parameters, paramType)
			}
		}
	}

	return &ProviderFunction{
		FunctionName: fn.Name.Name,
		Package:      pkg,
		ReturnType:   returnType,
		Parameters:   parameters,
		FilePath:     filePath,
	}
}

// hasErrorReturnType checks if a function returns error as the second return type
func (s *ASTScanner) hasErrorReturnType(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 2 {
		return false
	}

	// Check if second return type is error
	secondResult := fn.Type.Results.List[1]
	if ident, ok := secondResult.Type.(*ast.Ident); ok {
		return ident.Name == "error"
	}
	return false
}

// processTypeSpec analyzes type declarations for handler interfaces and implementations
func (s *ASTScanner) processTypeSpec(ts *ast.TypeSpec, pkg, filePath string, result *ScanResult) {
	typeName := ts.Name.Name

	switch t := ts.Type.(type) {
	case *ast.InterfaceType:
		// Check if this is a handler interface
		if s.isHandlerInterface(typeName, t) {
			methods := s.extractInterfaceMethods(t)
			result.Interfaces = append(result.Interfaces, HandlerInterface{
				InterfaceName: typeName,
				Package:       pkg,
				Methods:       methods,
				FilePath:      filePath,
			})
		}
	case *ast.StructType:
		// Check if this could be a handler implementation
		if s.isHandlerImplementation(typeName) {
			result.Implementations = append(result.Implementations, HandlerImplementation{
				StructName:    typeName,
				Package:       pkg,
				InterfaceName: "",         // Will be filled in during association
				Methods:       []string{}, // Will be filled when we find the methods
				FilePath:      filePath,
			})
		}
	}
}

// isHandlerInterface checks if an interface is likely a handler interface
func (s *ASTScanner) isHandlerInterface(name string, iface *ast.InterfaceType) bool {
	// Must be named "Handler"
	if name != "Handler" {
		return false
	}

	// Must have at least one method that looks like a handler
	for _, method := range iface.Methods.List {
		if funcType, ok := method.Type.(*ast.FuncType); ok {
			if s.hasFiberCtxParamInFunc(funcType) && s.returnsErrorInFunc(funcType) {
				return true
			}
		}
	}

	return false
}

// isHandlerImplementation checks if a struct is likely a handler implementation
func (s *ASTScanner) isHandlerImplementation(name string) bool {
	// Common patterns for implementation structs
	return name == "HandlerImpl" ||
		strings.HasSuffix(name, "Implementation") ||
		strings.HasSuffix(name, "Impl") ||
		(strings.HasSuffix(name, "Handler") && strings.Contains(name, "Impl"))
}

// extractInterfaceMethods extracts method names from an interface
func (s *ASTScanner) extractInterfaceMethods(iface *ast.InterfaceType) []string {
	var methods []string
	for _, method := range iface.Methods.List {
		if len(method.Names) > 0 {
			methods = append(methods, method.Names[0].Name)
		}
	}
	return methods
}

// hasFiberCtxParamInFunc checks if a function type has fiber.Ctx parameter
func (s *ASTScanner) hasFiberCtxParamInFunc(fn *ast.FuncType) bool {
	if fn.Params == nil || len(fn.Params.List) != 1 {
		return false
	}

	param := fn.Params.List[0]
	switch t := param.Type.(type) {
	case *ast.StarExpr:
		if sel, ok := t.X.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				return (ident.Name == "fiber" && sel.Sel.Name == "Ctx") ||
					(ident.Name == "gin" && sel.Sel.Name == "Context")
			}
		}
	}
	return false
}

// returnsErrorInFunc checks if a function type returns error
func (s *ASTScanner) returnsErrorInFunc(fn *ast.FuncType) bool {
	if fn.Results == nil || len(fn.Results.List) == 0 {
		return false
	}

	// Check last return type is error
	lastResult := fn.Results.List[len(fn.Results.List)-1]
	if ident, ok := lastResult.Type.(*ast.Ident); ok {
		return ident.Name == "error"
	}
	return false
}

// associateInterfacesWithImplementations links interfaces with their implementations
func (s *ASTScanner) associateInterfacesWithImplementations(result *ScanResult) {
	// For each implementation, try to find its corresponding interface
	for i := range result.Implementations {
		impl := &result.Implementations[i]

		// Look for an interface named "Handler" in the same package
		for _, iface := range result.Interfaces {
			if iface.Package == impl.Package && iface.InterfaceName == "Handler" {
				impl.InterfaceName = iface.InterfaceName
				break
			}
		}
	}

	// Now scan handlers again and create interface-based handlers
	s.createInterfaceBasedHandlers(result)
}

// createInterfaceBasedHandlers creates HandlerFunction entries for interface-based patterns
func (s *ASTScanner) createInterfaceBasedHandlers(result *ScanResult) {
	// Find handlers that are methods on implementation structs
	var newHandlers []HandlerFunction

	for _, handler := range result.Handlers {
		// Check if this handler is on an implementation struct
		for _, impl := range result.Implementations {
			if impl.Package == handler.Package && handler.HandlerName == impl.StructName {
				// Create a new interface-based handler
				newHandler := HandlerFunction{
					FunctionName:     handler.FunctionName,
					Package:          handler.Package,
					HandlerName:      impl.InterfaceName, // Use interface name
					ImplementerName:  impl.StructName,    // Store implementer name
					ReturnType:       handler.ReturnType,
					FilePath:         handler.FilePath,
					IsInterfaceBased: true,
				}
				newHandlers = append(newHandlers, newHandler)
			}
		}
	}

	// Add the interface-based handlers
	result.Handlers = append(result.Handlers, newHandlers...)
}

// Helper methods for AST analysis

func (s *ASTScanner) getReceiverTypeName(recv *ast.Field) string {
	switch t := recv.Type.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

func (s *ASTScanner) hasFiberCtxParam(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
		return false
	}

	param := fn.Type.Params.List[0]

	// Check for *fiber.Ctx or *gin.Context patterns
	switch t := param.Type.(type) {
	case *ast.StarExpr:
		if sel, ok := t.X.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				return (ident.Name == "fiber" && sel.Sel.Name == "Ctx") ||
					(ident.Name == "gin" && sel.Sel.Name == "Context")
			}
		}
	}

	return false
}

func (s *ASTScanner) returnsError(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return false
	}

	// Check last return type is error
	lastResult := fn.Type.Results.List[len(fn.Type.Results.List)-1]
	if ident, ok := lastResult.Type.(*ast.Ident); ok {
		return ident.Name == "error"
	}

	return false
}

func (s *ASTScanner) getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + s.getTypeString(t.X)
	case *ast.SelectorExpr:
		return s.getTypeString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + s.getTypeString(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", s.getTypeString(t.Key), s.getTypeString(t.Value))
	default:
		return ""
	}
}
