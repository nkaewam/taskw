package user

import (
	"fmt"
	"sync"
	"time"

	"github.com/example/ecommerce-api/internal/models"
	"github.com/google/uuid"
)

// Repository handles user data persistence
type Repository struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*models.User
}

// ProvideRepository creates a new user repository
func ProvideRepository() *Repository {
	return &Repository{
		users: make(map[uuid.UUID]*models.User),
	}
}

// Create creates a new user
func (r *Repository) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if email already exists
	for _, existingUser := range r.users {
		if existingUser.Email == user.Email {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
	}

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	r.users[user.ID] = user
	return nil
}

// GetByID retrieves a user by ID
func (r *Repository) GetByID(id uuid.UUID) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user with ID %s not found", id)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *Repository) GetByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user with email %s not found", email)
}

// GetAll retrieves all users
func (r *Repository) GetAll() ([]*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users, nil
}

// Update updates a user
func (r *Repository) Update(id uuid.UUID, updates map[string]interface{}) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user with ID %s not found", id)
	}

	// Create a copy to avoid modifying the original
	updatedUser := *user

	// Apply updates
	if email, ok := updates["email"].(string); ok {
		// Check if new email already exists
		for _, existingUser := range r.users {
			if existingUser.ID != id && existingUser.Email == email {
				return nil, fmt.Errorf("user with email %s already exists", email)
			}
		}
		updatedUser.Email = email
	}

	if firstName, ok := updates["first_name"].(string); ok {
		updatedUser.FirstName = firstName
	}

	if lastName, ok := updates["last_name"].(string); ok {
		updatedUser.LastName = lastName
	}

	updatedUser.UpdatedAt = time.Now()
	r.users[id] = &updatedUser

	return &updatedUser, nil
}

// Delete deletes a user
func (r *Repository) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[id]; !exists {
		return fmt.Errorf("user with ID %s not found", id)
	}

	delete(r.users, id)
	return nil
}
