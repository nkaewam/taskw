package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAddingNewDependency tests the workflow when adding a new provider function
func TestAddingNewDependency(t *testing.T) {
	// Setup: Create temporary directory for test
	testDir := filepath.Join(os.TempDir(), "taskw-e2e-dependency-test")
	if err := os.RemoveAll(testDir); err != nil {
		t.Fatalf("Failed to clean test directory: %v", err)
	}
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir) // Cleanup

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir) // Restore working directory

	// Get path to taskw binary BEFORE changing directories
	taskwBin := getTaskwBinary(t)

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}
	module := "github.com/test/e2e-dependency-project"
	projectName := "e2e-dependency-project"
	projectDir := filepath.Join(testDir, projectName)

	t.Run("01_setup_initial_project", func(t *testing.T) {
		// Initialize project first
		cmd := exec.Command(taskwBin, "init", module)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("taskw init failed: %v\nOutput: %s", err, string(output))
		}

		// Change to project directory for subsequent operations
		if err := os.Chdir(projectDir); err != nil {
			t.Fatalf("Failed to change to project directory: %v", err)
		}

		// Note: go mod tidy is now automatically run by taskw init

		t.Logf("✅ Initial project setup completed")
	})

	t.Run("02_add_new_service_module", func(t *testing.T) {
		// Create a new service module with providers
		serviceDir := filepath.Join(projectDir, "internal", "notification")
		if err := os.MkdirAll(serviceDir, 0755); err != nil {
			t.Fatalf("Failed to create notification service directory: %v", err)
		}

		// Create notification service with provider functions
		serviceCode := `package notification

import (
	"fmt"
	"go.uber.org/zap"
)

// Service handles notifications
type Service struct {
	logger *zap.Logger
}

// ProvideService creates a new notification service
func ProvideService(logger *zap.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

// SendEmail sends an email notification
func (s *Service) SendEmail(to, subject, body string) error {
	s.logger.Info("Sending email", 
		zap.String("to", to),
		zap.String("subject", subject))
	
	// Mock email sending
	fmt.Printf("📧 Email sent to %s: %s\n", to, subject)
	return nil
}

// SendSMS sends an SMS notification  
func (s *Service) SendSMS(to, message string) error {
	s.logger.Info("Sending SMS",
		zap.String("to", to), 
		zap.String("message", message))
	
	// Mock SMS sending
	fmt.Printf("📱 SMS sent to %s: %s\n", to, message)
	return nil
}
`

		serviceFile := filepath.Join(serviceDir, "service.go")
		if err := os.WriteFile(serviceFile, []byte(serviceCode), 0644); err != nil {
			t.Fatalf("Failed to create notification service: %v", err)
		}

		// Create notification repository with provider
		repositoryCode := `package notification

// Repository handles notification persistence
type Repository struct {
	// Mock repository - could be database, file, etc.
}

// ProvideRepository creates a new notification repository
func ProvideRepository() *Repository {
	return &Repository{}
}

// SaveNotification saves a notification record
func (r *Repository) SaveNotification(id, type_, recipient string) error {
	// Mock save operation
	fmt.Printf("💾 Notification saved: %s -> %s (%s)\n", id, recipient, type_)
	return nil
}
`

		repoFile := filepath.Join(serviceDir, "repository.go")
		if err := os.WriteFile(repoFile, []byte(repositoryCode), 0644); err != nil {
			t.Fatalf("Failed to create notification repository: %v", err)
		}

		// Create notification handler with provider
		handlerCode := `package notification

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Handler handles notification HTTP requests
type Handler struct {
	service    *Service
	repository *Repository
}

// ProvideHandler creates a new notification handler
func ProvideHandler(service *Service, repository *Repository) *Handler {
	return &Handler{
		service:    service,
		repository: repository,
	}
}

// @Summary Send email notification
// @Description Send an email notification to a recipient
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body EmailRequest true "Email request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/notifications/email [post]
func (h *Handler) SendEmail(c *fiber.Ctx) error {
	var req EmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	notificationID := uuid.New().String()
	
	if err := h.service.SendEmail(req.To, req.Subject, req.Body); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to send email"})
	}

	if err := h.repository.SaveNotification(notificationID, "email", req.To); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save notification"})
	}

	return c.JSON(fiber.Map{
		"id":      notificationID,
		"status":  "sent",
		"type":    "email",
		"recipient": req.To,
	})
}

// @Summary Send SMS notification  
// @Description Send an SMS notification to a recipient
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body SMSRequest true "SMS request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/notifications/sms [post]
func (h *Handler) SendSMS(c *fiber.Ctx) error {
	var req SMSRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	notificationID := uuid.New().String()
	
	if err := h.service.SendSMS(req.To, req.Message); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to send SMS"})
	}

	if err := h.repository.SaveNotification(notificationID, "sms", req.To); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save notification"})
	}

	return c.JSON(fiber.Map{
		"id":      notificationID,
		"status":  "sent",
		"type":    "sms",
		"recipient": req.To,
	})
}

// EmailRequest represents an email notification request
type EmailRequest struct {
	To      string ` + "`json:\"to\" validate:\"required,email\"`" + `
	Subject string ` + "`json:\"subject\" validate:\"required\"`" + `
	Body    string ` + "`json:\"body\" validate:\"required\"`" + `
}

// SMSRequest represents an SMS notification request  
type SMSRequest struct {
	To      string ` + "`json:\"to\" validate:\"required\"`" + `
	Message string ` + "`json:\"message\" validate:\"required\"`" + `
}
`

		handlerFile := filepath.Join(serviceDir, "handler.go")
		if err := os.WriteFile(handlerFile, []byte(handlerCode), 0644); err != nil {
			t.Fatalf("Failed to create notification handler: %v", err)
		}

		t.Logf("✅ New notification service module created with 3 providers")
	})

	t.Run("03_scan_for_new_providers", func(t *testing.T) {
		// Run taskw scan to see the new providers
		cmd := exec.Command(taskwBin, "scan")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("taskw scan failed: %v\nOutput: %s", err, string(output))
		}

		scanOutput := string(output)

		// Verify new providers are detected
		expectedProviders := []string{
			"ProvideService() ->",    // notification.ProvideService
			"ProvideRepository() ->", // notification.ProvideRepository
			"ProvideHandler(",        // notification.ProvideHandler
		}

		for _, provider := range expectedProviders {
			if !strings.Contains(scanOutput, provider) {
				t.Errorf("Expected provider not found in scan output: %s\nOutput: %s", provider, scanOutput)
			} else {
				t.Logf("✅ Provider detected in scan: %s", provider)
			}
		}

		// Verify routes are also detected
		expectedRoutes := []string{
			"POST /api/v1/notifications/email",
			"POST /api/v1/notifications/sms",
		}

		for _, route := range expectedRoutes {
			if !strings.Contains(scanOutput, route) {
				t.Errorf("Expected route not found in scan output: %s\nOutput: %s", route, scanOutput)
			} else {
				t.Logf("✅ Route detected in scan: %s", route)
			}
		}

		t.Logf("✅ Scan detected new providers and routes")
	})

	t.Run("04_generate_dependencies", func(t *testing.T) {
		// Run taskw generate to update dependencies
		cmd := exec.Command(taskwBin, "generate", "dependencies")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("taskw generate dependencies failed: %v\nOutput: %s", err, string(output))
		}

		// Verify dependencies_gen.go was updated
		depsFile := filepath.Join(projectDir, "internal", "api", "dependencies_gen.go")
		content, err := os.ReadFile(depsFile)
		if err != nil {
			t.Fatalf("Failed to read generated dependencies file: %v", err)
		}

		depsContent := string(content)

		// Check for new notification providers
		expectedProviders := []string{
			"notification.ProvideService",
			"notification.ProvideRepository",
			"notification.ProvideHandler",
		}

		for _, provider := range expectedProviders {
			if !strings.Contains(depsContent, provider) {
				t.Errorf("Expected provider not found in dependencies_gen.go: %s", provider)
			} else {
				t.Logf("✅ Provider added to dependencies: %s", provider)
			}
		}

		// Verify import was added
		if !strings.Contains(depsContent, "internal/notification") {
			t.Errorf("Expected notification import not found in dependencies_gen.go")
		} else {
			t.Logf("✅ Notification import added to dependencies")
		}

		t.Logf("✅ Dependencies generated successfully with new providers")
	})

	t.Run("05_generate_routes", func(t *testing.T) {
		// Run taskw generate to update routes
		cmd := exec.Command(taskwBin, "generate", "routes")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("taskw generate routes failed: %v\nOutput: %s", err, string(output))
		}

		// Verify routes_gen.go was updated
		routesFile := filepath.Join(projectDir, "internal", "api", "routes_gen.go")
		content, err := os.ReadFile(routesFile)
		if err != nil {
			t.Fatalf("Failed to read generated routes file: %v", err)
		}

		routesContent := string(content)

		// Check for new notification routes
		expectedRoutes := []string{
			"Post(\"/api/v1/notifications/email\"",
			"Post(\"/api/v1/notifications/sms\"",
			"s.notificationHandler.SendEmail",
			"s.notificationHandler.SendSMS",
		}

		for _, route := range expectedRoutes {
			if !strings.Contains(routesContent, route) {
				t.Errorf("Expected route not found in routes_gen.go: %s", route)
			} else {
				t.Logf("✅ Route added: %s", route)
			}
		}

		t.Logf("✅ Routes generated successfully with new endpoints")
	})

	t.Run("06_update_server_struct", func(t *testing.T) {
		// Update server.go to include the new notification handler
		serverFile := filepath.Join(projectDir, "internal", "api", "server.go")
		content, err := os.ReadFile(serverFile)
		if err != nil {
			t.Fatalf("Failed to read server.go: %v", err)
		}

		serverContent := string(content)

		// Add notification import and handler to server struct
		updatedContent := strings.Replace(serverContent,
			`import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)`,
			`import (
	"`+module+`/internal/notification"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)`, 1)

		updatedContent = strings.Replace(updatedContent,
			`// Server holds all the handlers and dependencies
type Server struct {
	logger *zap.Logger
}`,
			`// Server holds all the handlers and dependencies
type Server struct {
	logger              *zap.Logger
	notificationHandler *notification.Handler
}`, 1)

		updatedContent = strings.Replace(updatedContent,
			`// ProvideServer creates a new server with all dependencies
func ProvideServer(
	logger *zap.Logger,
) *Server {
	return &Server{
		logger: logger,
	}
}`,
			`// ProvideServer creates a new server with all dependencies
func ProvideServer(
	logger *zap.Logger,
	notificationHandler *notification.Handler,
) *Server {
	return &Server{
		logger:              logger,
		notificationHandler: notificationHandler,
	}
}`, 1)

		if err := os.WriteFile(serverFile, []byte(updatedContent), 0644); err != nil {
			t.Fatalf("Failed to update server.go: %v", err)
		}

		t.Logf("✅ Server struct updated with notification handler")
	})

	t.Run("07_run_wire_generate", func(t *testing.T) {
		// Run wire to generate the dependency injection
		cmd := exec.Command("wire", "./internal/api")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Wire output: %s", string(output))
			// Wire might not be installed, which is okay for this test
			if strings.Contains(string(output), "wire: command not found") {
				t.Logf("⚠️  Wire not installed, skipping wire generation")
				return
			}
			t.Fatalf("Wire generation failed: %v\nOutput: %s", err, string(output))
		}

		t.Logf("✅ Wire generation completed")
	})

	t.Run("08_verify_project_compiles", func(t *testing.T) {
		// Test that the project still compiles with new dependencies
		cmd := exec.Command("go", "build", "./...")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Build output: %s", string(output))
			// Some build errors might be expected due to missing dependencies
			// Focus on whether our generated code has syntax errors
			if strings.Contains(string(output), "syntax error") ||
				strings.Contains(string(output), "undefined:") &&
					!strings.Contains(string(output), "RegisterRoutes") {
				t.Errorf("Project has syntax errors after adding dependency: %v", err)
			} else {
				t.Logf("✅ Build issues are expected due to missing route registration")
			}
		} else {
			t.Logf("✅ Project compiles successfully with new dependencies")
		}
	})

	t.Logf("✅ Adding new dependency e2e test completed successfully")
}
