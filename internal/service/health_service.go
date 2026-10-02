package service

import (
	"context"
	"time"

	"github.com/niviten/heavy/internal/dto"
)

const healthCheckTimeout = 2 * time.Second

// HealthRepository is the persistence behavior required by the service.
type HealthRepository interface {
	Ping(context.Context) error
}

// HealthService checks application dependencies.
type HealthService struct {
	repository HealthRepository
}

func NewHealthService(repository HealthRepository) *HealthService {
	return &HealthService{repository: repository}
}

// Check returns a healthy result only when MySQL responds before the timeout.
func (s *HealthService) Check(ctx context.Context) (dto.HealthResponse, error) {
	checkCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	if err := s.repository.Ping(checkCtx); err != nil {
		return dto.HealthResponse{Status: "unhealthy", Database: "down"}, err
	}

	return dto.HealthResponse{Status: "ok", Database: "up"}, nil
}
