package repo

import "context"

// HealthRepository defines the interface for health repository operations.
type HealthRepository interface {
	Ping(ctx context.Context) error
}

// healthRepository is an in-memory implementation for unit tests.
type healthRepository struct{}

// NewHealthRepository creates a new in-memory health repository.
func NewHealthRepository() HealthRepository {
	return &healthRepository{}
}

// Ping checks the repository connection.
func (r *healthRepository) Ping(ctx context.Context) error {
	return nil
}
