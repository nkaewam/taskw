package health

import (
	"time"
)

// HealthStatus represents the health status of the system
type HealthStatus struct {
	Status  string        `json:"status"`
	Service string        `json:"service"`
	Uptime  time.Duration `json:"uptime"`
}

// Service handles health business logic
type Service struct {
	repo      *Repository
	startTime time.Time
}

// ProvideService creates a new health service
func ProvideService(repo *Repository) *Service {
	return &Service{
		repo:      repo,
		startTime: time.Now(),
	}
}

// GetHealth returns the current health status of the system
func (s *Service) GetHealth() *HealthStatus {
	// Check system components status
	isHealthy := s.repo.CheckSystemHealth()

	status := "healthy"
	if !isHealthy {
		status = "unhealthy"
	}

	return &HealthStatus{
		Status:  status,
		Service: "ecommerce-api",
		Uptime:  time.Since(s.startTime),
	}
}
