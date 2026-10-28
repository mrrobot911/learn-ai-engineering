package service

import "context"

// HealthService defines the interface for health checking.
type HealthService interface {
	Check(ctx context.Context) error
}

// healthService implements HealthService.
type healthService struct{}

// NewHealthService creates a new health service.
func NewHealthService() HealthService {
	return &healthService{}
}

// Check performs a health check.
func (s *healthService) Check(ctx context.Context) error {
	return nil
}
