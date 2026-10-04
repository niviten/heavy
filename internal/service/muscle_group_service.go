package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/niviten/heavy/internal/model"
)

var ErrMuscleGroupNotFound = errors.New("muscle group not found")

// MuscleGroupRepository is the persistence behavior required by the service.
type MuscleGroupRepository interface {
	GetAll(context.Context) ([]model.MuscleGroup, error)
}

// MuscleGroupService lazily loads the read-only muscle_groups table and keeps
// indexed, immutable copies in memory for subsequent callers.
type MuscleGroupService struct {
	repository MuscleGroupRepository

	mu     sync.RWMutex
	loaded bool
	all    []model.MuscleGroup
	byID   map[int64]model.MuscleGroup
	byName map[string]model.MuscleGroup
}

func NewMuscleGroupService(repository MuscleGroupRepository) *MuscleGroupService {
	return &MuscleGroupService{repository: repository}
}

func (s *MuscleGroupService) GetAll(ctx context.Context) ([]model.MuscleGroup, error) {
	if err := s.load(ctx); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]model.MuscleGroup, len(s.all))
	copy(result, s.all)
	return result, nil
}

func (s *MuscleGroupService) GetByID(ctx context.Context, id int64) (model.MuscleGroup, error) {
	if err := s.load(ctx); err != nil {
		return model.MuscleGroup{}, err
	}

	s.mu.RLock()
	muscleGroup, ok := s.byID[id]
	s.mu.RUnlock()
	if !ok {
		return model.MuscleGroup{}, ErrMuscleGroupNotFound
	}

	return muscleGroup, nil
}

func (s *MuscleGroupService) GetByName(ctx context.Context, name string) (model.MuscleGroup, error) {
	if err := s.load(ctx); err != nil {
		return model.MuscleGroup{}, err
	}

	s.mu.RLock()
	muscleGroup, ok := s.byName[normalizeMuscleGroupName(name)]
	s.mu.RUnlock()
	if !ok {
		return model.MuscleGroup{}, ErrMuscleGroupNotFound
	}

	return muscleGroup, nil
}

func (s *MuscleGroupService) load(ctx context.Context) error {
	s.mu.RLock()
	loaded := s.loaded
	s.mu.RUnlock()
	if loaded {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loaded {
		return nil
	}

	muscleGroups, err := s.repository.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("load muscle groups: %w", err)
	}

	s.all = make([]model.MuscleGroup, len(muscleGroups))
	copy(s.all, muscleGroups)
	s.byID = make(map[int64]model.MuscleGroup, len(muscleGroups))
	s.byName = make(map[string]model.MuscleGroup, len(muscleGroups))
	for _, muscleGroup := range muscleGroups {
		s.byID[muscleGroup.ID] = muscleGroup
		s.byName[normalizeMuscleGroupName(muscleGroup.Name)] = muscleGroup
	}
	s.loaded = true

	return nil
}

func normalizeMuscleGroupName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
