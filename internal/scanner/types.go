package scanner

// HandlerFunction represents a Fiber handler function found in the codebase
type HandlerFunction struct {
	FunctionName string // e.g., "GetUser"
	Package      string // e.g., "user"
	HandlerName  string // e.g., "UserHandler"
	ReturnType   string // Always "error" for Fiber handlers
	FilePath     string // Path to the file containing this handler
}

// RouteMapping represents a @Router annotation mapping
type RouteMapping struct {
	MethodName string // e.g., "GetUser"
	Path       string // e.g., "/users/:id"
	HTTPMethod string // e.g., "GET", "POST", "PUT", "DELETE"
	HandlerRef string // e.g., "userHandler.GetUser"
	Package    string // Package name for import resolution
}

// ProviderFunction represents a Wire provider function
type ProviderFunction struct {
	FunctionName string   // e.g., "ProvideUserService"
	Package      string   // e.g., "user"
	ReturnType   string   // e.g., "*UserService"
	Parameters   []string // Parameter types for dependency resolution
	FilePath     string   // Path to the file containing this provider
}

// ScanResult aggregates all scanning results
type ScanResult struct {
	Handlers  []HandlerFunction
	Routes    []RouteMapping
	Providers []ProviderFunction
	Errors    []ScanError
}

// ScanError represents an error encountered during scanning
type ScanError struct {
	FilePath string
	Line     int
	Message  string
	Type     string // "handler", "route", "provider"
}
