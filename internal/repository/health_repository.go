// Package repository contains all database operations.
package repository

import "context"

// DatabasePinger describes the database operation required for a health check.
type DatabasePinger interface {
	PingContext(context.Context) error
}

// HealthRepository checks database availability.
type HealthRepository struct {
	database DatabasePinger
}

func NewHealthRepository(database DatabasePinger) *HealthRepository {
	return &HealthRepository{database: database}
}

func (r *HealthRepository) Ping(ctx context.Context) error {
	return r.database.PingContext(ctx)
}
