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
		Handlers:  []HandlerFunction{},
		Routes:    []RouteMapping{},
		Providers: []ProviderFunction{},
		Errors:    []ScanError{},
	}

	packageName := node.Name.Name

	// Walk the AST to find functions
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			s.processFuncDecl(x, packageName, filePath, result)
		}
		return true
	})

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

	// Check receiver type: should be *SomeHandler
	recv := fn.Recv.List[0]
	handlerName := s.getReceiverTypeName(recv)
	if handlerName == "" || !strings.HasSuffix(handlerName, "Handler") {
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
func (s *ASTScanner) extractRoute(fn *ast.FuncDecl, handler HandlerFunction) *RouteMapping {
	if fn.Doc == nil {
		return nil
	}

	// Regex to parse @Router /path [method]
	routerRegex := regexp.MustCompile(`@Router\s+([^\s]+)\s+\[([^\]]+)\]`)

	for _, comment := range fn.Doc.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		if matches := routerRegex.FindStringSubmatch(text); matches != nil {
			path := matches[1]
			method := strings.ToUpper(matches[2])

			return &RouteMapping{
				MethodName: fn.Name.Name,
				Path:       path,
				HTTPMethod: method,
				HandlerRef: fmt.Sprintf("%s.%s", strings.ToLower(handler.HandlerName[:len(handler.HandlerName)-7]), handler.FunctionName), // Remove "Handler" suffix
				Package:    handler.Package,
			}
		}
	}

	return nil
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
