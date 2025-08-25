# Taskw End-to-End Tests

This directory contains comprehensive end-to-end tests for Taskw that verify the complete workflows developers will experience when using the tool.

## Test Cases

### 1. Project Initialization (`01_init_test.go`)

**Scenario**: Developer creates a new Taskw project from scratch

**Test Steps**:

1. Run `taskw init` with a Go module path
2. Verify all scaffolded files are created correctly
3. Check go.mod has correct module and dependencies
4. Validate taskw.yaml configuration
5. Verify health handler template content
6. Check .taskwignore patterns
7. Ensure `go mod tidy` works
8. Verify project has correct structure and compiles

**Expected Outcome**: A fully functional Taskw project ready for development

### 2. Adding New Dependency (`02_dependency_test.go`)

**Scenario**: Developer adds a new service with provider functions

**Test Steps**:

1. Start with initialized project
2. Add new notification service module with multiple providers
3. Run `taskw scan` to detect new providers
4. Generate dependencies to include new providers
5. Generate routes for new handlers
6. Update server struct to include new handlers
7. Run wire generation
8. Verify project still compiles

**Expected Outcome**: New providers are automatically detected and included in dependency injection

### 3. Adding New Route (`03_route_test.go`)

**Scenario**: Developer adds new handler methods with @Router annotations

**Test Steps**:

1. Start with project + basic user service
2. Create initial handler with 2 routes
3. Generate initial routes
4. Add 4 new handler methods with @Router annotations
5. Scan to detect new routes
6. Regenerate routes
7. Verify route ordering and conflict resolution
8. Update server to include handlers
9. Verify project builds with new routes

**Expected Outcome**: New routes are automatically detected and added to route registration

## Running the Tests

### Prerequisites

```bash
# Build Taskw binary
cd /path/to/taskw
go build -o bin/taskw cmd/taskw/main.go

# Ensure Go and basic tools are installed
go version  # Should be Go 1.21+
```

### Run All E2E Tests

```bash
cd test/e2e
go test -v ./...
```

### Run Specific Test

```bash
cd test/e2e
go test -v -run TestProjectInitialization
go test -v -run TestAddingNewDependency
go test -v -run TestAddingNewRoute
```

### Run with Verbose Output

```bash
cd test/e2e
go test -v -run TestProjectInitialization 2>&1 | tee init_test.log
```

## Test Environment

Each test creates its own temporary directory under `/tmp/taskw-e2e-*-test` and cleans up automatically. Tests are designed to be:

- **Isolated**: Each test runs in its own directory
- **Self-contained**: No dependencies on external services
- **Fast**: Focus on critical path verification
- **Comprehensive**: Cover the major user workflows

## Expected Test Results

### Successful Run Example

```bash
=== RUN   TestProjectInitialization
=== RUN   TestProjectInitialization/01_initialize_project
    01_init_test.go:45: ✅ Project directory created: /tmp/taskw-e2e-init-test/e2e-init-project
=== RUN   TestProjectInitialization/02_verify_scaffolded_files
    01_init_test.go:60: ✅ File exists: cmd/server/main.go
    01_init_test.go:60: ✅ File exists: internal/api/server.go
    01_init_test.go:60: ✅ File exists: internal/api/wire.go
    # ... more verification steps
=== RUN   TestProjectInitialization/08_project_compiles
    01_init_test.go:180: ✅ Build issues are related to missing RegisterRoutes method (expected)
--- PASS: TestProjectInitialization (2.34s)
```

## Troubleshooting

### Test Failures

**"Could not find taskw binary"**

```bash
# Ensure taskw is built
go build -o bin/taskw cmd/taskw/main.go
```

**"go mod tidy failed"**

```bash
# Check Go version and module setup
go version
go env GOMOD
```

**"Wire generation failed"**

```bash
# Install wire (optional for tests)
go install github.com/google/wire/cmd/wire@latest
```

### Debug Mode

Set environment variable for more detailed output:

```bash
export TASKW_E2E_DEBUG=1
go test -v ./...
```

## Adding New E2E Tests

When adding new test scenarios:

1. **Create new file**: `XX_feature_test.go`
2. **Follow naming**: `TestFeatureName`
3. **Use subtests**: Break into logical steps
4. **Clean up**: Use `defer os.RemoveAll(testDir)`
5. **Verify thoroughly**: Check generated files, compilation, functionality
6. **Document**: Add to this README

### Test Template

```go
func TestNewFeature(t *testing.T) {
    // Setup
    testDir := filepath.Join(os.TempDir(), "taskw-e2e-feature-test")
    defer os.RemoveAll(testDir)

    t.Run("01_setup", func(t *testing.T) {
        // Test setup
    })

    t.Run("02_main_functionality", func(t *testing.T) {
        // Core test logic
    })

    t.Run("03_verification", func(t *testing.T) {
        // Verify expected outcomes
    })

    t.Logf("✅ Feature e2e test completed successfully")
}
```

This ensures consistency and maintainability across all e2e tests.
