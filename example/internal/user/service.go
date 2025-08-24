package user

import (
	"fmt"

	"github.com/example/ecommerce-api/internal/models"
	"github.com/google/uuid"
)

// Service handles user business logic
type Service struct {
	repo *Repository
}

// ProvideService creates a new user service
func ProvideService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateUser creates a new user
func (s *Service) CreateUser(req *models.CreateUserRequest) (*models.UserResponse, error) {
	// Validate business rules
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, err
	}

	user := &models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(id uuid.UUID) (*models.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// GetUsers retrieves all users
func (s *Service) GetUsers() ([]*models.UserResponse, error) {
	users, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	responses := make([]*models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.toUserResponse(user)
	}

	return responses, nil
}

// UpdateUser updates a user
func (s *Service) UpdateUser(id uuid.UUID, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	// Validate business rules
	if err := s.validateUpdateUserRequest(req); err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}

	user, err := s.repo.Update(id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(id uuid.UUID) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (s *Service) GetUserByEmail(email string) (*models.UserResponse, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return s.toUserResponse(user), nil
}

// validateCreateUserRequest validates create user request
func (s *Service) validateCreateUserRequest(req *models.CreateUserRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.FirstName == "" {
		return fmt.Errorf("first name is required")
	}
	if req.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	if len(req.FirstName) < 2 || len(req.FirstName) > 50 {
		return fmt.Errorf("first name must be between 2 and 50 characters")
	}
	if len(req.LastName) < 2 || len(req.LastName) > 50 {
		return fmt.Errorf("last name must be between 2 and 50 characters")
	}
	return nil
}

// validateUpdateUserRequest validates update user request
func (s *Service) validateUpdateUserRequest(req *models.UpdateUserRequest) error {
	if req.FirstName != nil && (len(*req.FirstName) < 2 || len(*req.FirstName) > 50) {
		return fmt.Errorf("first name must be between 2 and 50 characters")
	}
	if req.LastName != nil && (len(*req.LastName) < 2 || len(*req.LastName) > 50) {
		return fmt.Errorf("last name must be between 2 and 50 characters")
	}
	return nil
}

// toUserResponse converts a User model to UserResponse
func (s *Service) toUserResponse(user *models.User) *models.UserResponse {
	return &models.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	}
}
