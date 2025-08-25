package ui

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Service handles user interface operations like spinners and prompts
type Service interface {
	// ShowSpinner displays a spinner with a message and returns a stop function
	ShowSpinner(message string) func(completedMessage string)
	// PromptForModule interactively prompts for a Go module path
	PromptForModule() (string, error)
}

// service implements Service interface
type service struct{}

// ProvideUIService creates a new UI service
// @Provider
func ProvideUIService() Service {
	return &service{}
}

// ShowSpinner displays a spinner with a message and returns a stop function
func (s *service) ShowSpinner(message string) func(completedMessage string) {
	spinner := NewSpinner()
	spinner.Start(message)
	return func(completedMessage string) {
		spinner.Stop(completedMessage)
	}
}

// PromptForModule interactively prompts for a Go module path
func (s *service) PromptForModule() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("🚀 Let's create a new Taskw project!")
	fmt.Println()

	for {
		fmt.Print("Enter Go module path (e.g., github.com/username/my-awesome-api): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		module := strings.TrimSpace(input)
		if module == "" {
			fmt.Println("❌ Module path cannot be empty. Please try again.")
			continue
		}

		if err := ValidateModule(module); err != nil {
			fmt.Printf("❌ %v Please try again.\n", err)
			continue
		}

		projectName := ExtractProjectName(module)
		fmt.Printf("✅ Great! Creating project '%s' with module '%s'\n", projectName, module)
		return module, nil
	}
}

// Spinner handles animated loading indicators
type Spinner struct {
	chars   []string
	delay   time.Duration
	done    chan bool
	mu      sync.Mutex
	stopped bool // Track if spinner has been stopped
}

func NewSpinner() *Spinner {
	return &Spinner{
		chars: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		delay: 100 * time.Millisecond,
		done:  make(chan bool, 1), // Make the channel buffered to prevent deadlock
	}
}

func (s *Spinner) Start(message string) {
	go func() {
		i := 0
		for {
			select {
			case <-s.done:
				return
			default:
				s.mu.Lock()
				fmt.Printf("\r%s %s", s.chars[i%len(s.chars)], message)
				s.mu.Unlock()
				i++
				time.Sleep(s.delay)
			}
		}
	}()
}

func (s *Spinner) Stop(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prevent multiple calls to Stop
	if s.stopped {
		return
	}
	s.stopped = true

	// Send stop signal (non-blocking with buffered channel)
	select {
	case s.done <- true:
	default:
		// Channel already has a value, that's fine
	}

	fmt.Printf("\r✔ %s\n", message)
}

// ValidateModule validates that the module path is a proper Go module format
func ValidateModule(module string) error {
	// Basic module format validation
	if !strings.Contains(module, "/") {
		return fmt.Errorf("module must contain at least one '/' (e.g., github.com/user/project)")
	}

	// Check for valid module path format
	modulePattern := regexp.MustCompile(`^[a-zA-Z0-9.-]+(\.[a-zA-Z]{2,})?(/[a-zA-Z0-9._-]+)+$`)
	if !modulePattern.MatchString(module) {
		return fmt.Errorf("invalid module format. Use format like: github.com/user/project-name")
	}

	// Extract and validate project name (last part of module)
	projectName := ExtractProjectName(module)
	return ValidateProjectName(projectName)
}

// ExtractProjectName extracts the project name from a module path
func ExtractProjectName(module string) string {
	parts := strings.Split(module, "/")
	return parts[len(parts)-1]
}

// ValidateProjectName validates that the project name follows slug-case format
func ValidateProjectName(name string) error {
	// Check for slug-case format: lowercase letters, numbers, and hyphens only
	// Cannot start or end with hyphen, cannot have consecutive hyphens
	slugPattern := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

	if !slugPattern.MatchString(name) {
		return fmt.Errorf("project name (last part of module) must be in slug-case (lowercase letters, numbers, and hyphens only, e.g., 'my-api')")
	}

	// Additional validation rules
	if len(name) < 2 {
		return fmt.Errorf("project name must be at least 2 characters long")
	}

	if len(name) > 50 {
		return fmt.Errorf("project name must be no longer than 50 characters")
	}

	// Check for reserved names
	reservedNames := []string{
		"api", "app", "main", "src", "lib", "bin", "cmd", "internal",
		"pkg", "test", "tests", "doc", "docs", "build", "dist",
	}

	for _, reserved := range reservedNames {
		if name == reserved {
			return fmt.Errorf("'%s' is a reserved name, please choose a different project name", name)
		}
	}

	return nil
}
