# Taskw MVP Development TODOs

## Project Status

Taskw is a Go CLI tool that generates Fiber routes and Wire dependency injection code from annotations. The goal is to create an MVP that can launch and provide real value to Go developers using Fiber + Wire + Swaggo.

**Current State:**

- ✅ Config system (mostly complete with Viper)
- ⚠️ CLI interface (placeholder only)
- ❌ Scanner system (not implemented)
- ❌ Generator system (empty files)
- ❌ Templates (not created)
- ❌ Tests (none)

## Phase 1: Core Functionality (Essential for MVP)

### 1.1 CLI Foundation

- [x] **CLI Commands Structure** - Replace placeholder main.go with proper CLI
  - Commands: `init`, `generate [all|routes|deps]`, `scan`
  - Use cobra or simple flag package
  - Proper error handling and help messages

### 1.2 Scanner Implementation

- [x] **Route Scanner** - Extract @Router annotations from handler.go files
  - Parse `@Router /path [method]` comments
  - Identify handler functions with proper signatures `func (h *Handler) Method(c *fiber.Ctx) error`
  - Extract package information for imports
- [x] **Provider Scanner** - Find Provide\* functions across codebase
  - Locate functions starting with `Provide*`
  - Extract return types and parameter dependencies
  - Map packages for proper import generation

### 1.3 Code Generation

- [x] **Route Generator** - Generate Fiber route registration code ✅

  - Create `routes_gen.go` with proper package and imports ✅
  - Generate route registration: `app.Get("/path", handler.Method)` ✅
  - Handle different HTTP methods (GET, POST, PUT, DELETE, etc.) ✅

- [x] **Dependency Generator** - Generate Wire provider sets ✅
  - Create `dependencies_gen.go` with Wire imports ✅
  - Generate `var ProviderSet = wire.NewSet(...)` with all Provide\* functions ✅
  - Proper package imports for all providers ✅

### 1.4 Template System

- [x] **Go Templates** - Create templates for code generation
  - `templates/routes.tmpl` for Fiber route registration
  - `templates/dependencies.tmpl` for Wire provider sets
  - Templates should handle imports automatically

### 1.5 File Management

- [ ] **Safe File Writing** - Implement proper file output
  - Use `go/format` to ensure generated code is properly formatted
  - Create backup files before overwriting
  - Ensure output directories exist

## Phase 2: Quality & Reliability

### 2.1 Error Handling

- [ ] **Comprehensive Error Messages** - User-friendly errors with suggestions
  - Config validation errors with specific fixes
  - Scanner errors when annotations are malformed
  - Generator errors with context about what failed

### 2.2 Validation

- [ ] **Config Validation** - Ensure taskw.yaml is properly configured

  - Validate scan directories exist
  - Check output directory is writable
  - Verify Go module format

- [ ] **Scanner Validation** - Validate discovered code patterns
  - Ensure handler functions have correct signatures
  - Validate @Router annotation syntax
  - Check for duplicate routes

### 2.3 Import Resolution

- [ ] **Smart Import Management** - Automatically resolve import paths
  - Generate correct import statements based on scanned packages
  - Handle relative vs absolute imports correctly
  - Avoid duplicate imports

## Phase 3: Testing (Critical for Production)

### 3.1 Unit Tests

- [ ] **Config Tests** - Test configuration loading and validation

  - Default config creation
  - YAML parsing and saving
  - Go module detection

- [ ] **Scanner Tests** - Test code parsing with sample files

  - Route extraction from various @Router formats
  - Provider function detection
  - Edge cases and malformed code

- [ ] **Generator Tests** - Test code generation with known inputs
  - Route generation produces expected Fiber code
  - Dependency generation produces expected Wire code
  - Template rendering works correctly

### 3.2 Integration Tests

- [ ] **End-to-End Test** - Full workflow testing
  - Create sample Go project with handlers and providers
  - Run Taskw generate command
  - Verify generated code compiles and runs
  - Test with different project structures

### 3.3 Example Project

- [ ] **Complete Demo** - Working example in `examples/` directory
  - Simple API with user/product handlers
  - Proper @Router annotations
  - Provide\* functions for DI
  - Generated code that actually works
  - README showing how to run the example

## Phase 4: Distribution & Documentation

### 4.1 Build Automation

- [ ] **Build System** - Automate building and releasing
  - Makefile or Taskfile for common tasks
  - Cross-platform binary builds
  - Version management

### 4.2 CI/CD

- [ ] **GitHub Actions** - Automated testing and releasing
  - Run tests on PRs
  - Build binaries on release tags
  - Publish to GitHub releases

### 4.3 Documentation Polish

- [ ] **README Updates** - Ensure all examples work
  - Verify installation instructions
  - Test all code examples
  - Add troubleshooting section

### 4.4 Pre-Launch Validation

- [ ] **Manual Testing** - Test on fresh projects
  - Install on clean system
  - Test with various Go project structures
  - Verify user experience is smooth

## Phase 5: Future Enhancements (Post-MVP)

These are nice-to-have features that can be added after the MVP launch:

- [ ] **Watch Mode** - Auto-regenerate on file changes (`taskw watch`)
- [ ] **Verbose Mode** - Debug output for troubleshooting
- [ ] **Dry Run Mode** - Show what would be generated without writing files
- [ ] **Custom Templates** - Allow users to customize generated code
- [ ] **Multiple Output Formats** - Support different project structures
- [ ] **Configuration Validation** - More sophisticated config checking

## Success Criteria for MVP Launch

### Must Have:

1. ✅ `taskw init` creates working configuration
2. ✅ `taskw generate` produces compilable Go code
3. ✅ Generated routes work with Fiber
4. ✅ Generated dependencies work with Wire
5. ✅ Basic error handling prevents crashes
6. ✅ One complete working example
7. ✅ Installation via `go install` works

### Should Have:

1. ✅ Unit tests cover core functionality
2. ✅ Integration test proves end-to-end workflow
3. ✅ Good error messages guide users to solutions
4. ✅ Documentation is complete and tested

### Could Have:

1. Advanced CLI features (verbose, dry-run)
2. Watch mode
3. Custom templates
4. Multiple framework support

## Development Priority

**Week 1-2: Core Implementation**

- CLI structure, scanners, generators, templates

**Week 3: Quality & Testing**

- Error handling, validation, unit tests

**Week 4: Integration & Polish**

- End-to-end tests, example project, documentation

**Target: MVP launch within 4 weeks**

## Notes

- Keep it simple - focus on the 80% use case
- Don't over-engineer - we can improve post-launch
- Prioritize working code over perfect code
- User experience matters more than feature completeness
