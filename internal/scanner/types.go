package scanner

// HandlerFunction represents a Fiber handler function found in the codebase
type HandlerFunction struct {
	FunctionName     string // e.g., "GetUser"
	Package          string // e.g., "user"
	FullPackagePath  string // e.g., "domain/user" (complete package path from file system)
	HandlerName      string // e.g., "UserHandler" (interface name if using interface pattern)
	ImplementerName  string // e.g., "HandlerImpl" (only for interface pattern)
	ReturnType       string // Always "error" for Fiber handlers
	FilePath         string // Path to the file containing this handler
	IsInterfaceBased bool   // true if this handler uses interface + implementation pattern
}

// RouteMapping represents a @Router annotation mapping
type RouteMapping struct {
	MethodName      string // e.g., "GetUser"
	Path            string // e.g., "/users/:id"
	HTTPMethod      string // e.g., "GET", "POST", "PUT", "DELETE"
	HandlerRef      string // e.g., "userHandler.GetUser"
	Package         string // Package name for import resolution
	FullPackagePath string // e.g., "domain/user" (complete package path from file system)
}

// ProviderFunction represents a Wire provider function
type ProviderFunction struct {
	FunctionName string   // e.g., "ProvideUserService"
	Package      string   // e.g., "user"
	ReturnType   string   // e.g., "*UserService"
	Parameters   []string // Parameter types for dependency resolution
	FilePath     string   // Path to the file containing this provider
}

// HandlerInterface represents a handler interface definition
type HandlerInterface struct {
	InterfaceName string   // e.g., "Handler"
	Package       string   // e.g., "user"
	Methods       []string // e.g., ["GetUser", "CreateUser"]
	FilePath      string   // Path to the file containing this interface
}

// HandlerImplementation represents a handler implementation struct
type HandlerImplementation struct {
	StructName    string   // e.g., "HandlerImpl"
	Package       string   // e.g., "user"
	InterfaceName string   // e.g., "Handler" (the interface it implements)
	Methods       []string // e.g., ["GetUser", "CreateUser"]
	FilePath      string   // Path to the file containing this struct
}

// ScanResult aggregates all scanning results
type ScanResult struct {
	Handlers        []HandlerFunction
	Routes          []RouteMapping
	Providers       []ProviderFunction
	Interfaces      []HandlerInterface      // Handler interfaces found
	Implementations []HandlerImplementation // Handler implementations found
	Errors          []ScanError
}

// ScanError represents an error encountered during scanning
type ScanError struct {
	FilePath string
	Line     int
	Message  string
	Type     string // "handler", "route", "provider"
}
