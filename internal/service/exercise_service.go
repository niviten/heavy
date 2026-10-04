package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/niviten/heavy/internal/model"
	"github.com/niviten/heavy/internal/repository"
)

const (
	DefaultExercisePageSize = 20
	MaxExercisePageSize     = 100
)

var (
	ErrExerciseNotFound     = repository.ErrExerciseNotFound
	ErrExerciseNameConflict = repository.ErrExerciseNameConflict
	ErrInvalidExerciseName  = errors.New("exercise name must contain between 1 and 100 characters")
	ErrInvalidMuscleGroupID = errors.New("muscle group id must be a positive integer")
	ErrEmptyExerciseUpdate  = errors.New("at least one exercise field must be provided")
)

// ExerciseRepository is the persistence behavior required by the service.
type ExerciseRepository interface {
	GetAll(context.Context, model.ExerciseQuery) ([]model.Exercise, int64, error)
	GetByMuscleGroup(context.Context, int64, model.ExerciseQuery) ([]model.Exercise, int64, error)
	GetByID(context.Context, int64) (model.Exercise, error)
	Create(context.Context, string, int64) (model.Exercise, error)
	Update(context.Context, int64, *string, *int64) (model.Exercise, error)
	SoftDelete(context.Context, int64) (bool, error)
}

// MuscleGroupLookup verifies referenced muscle groups.
type MuscleGroupLookup interface {
	GetByID(context.Context, int64) (model.MuscleGroup, error)
}

type ExerciseService struct {
	repository   ExerciseRepository
	muscleGroups MuscleGroupLookup
}

func NewExerciseService(
	repository ExerciseRepository,
	muscleGroups MuscleGroupLookup,
) *ExerciseService {
	return &ExerciseService{repository: repository, muscleGroups: muscleGroups}
}

func (s *ExerciseService) GetAll(
	ctx context.Context,
	query model.ExerciseQuery,
) (model.ExercisePage, error) {
	query = normalizeExerciseQuery(query)
	items, total, err := s.repository.GetAll(ctx, query)
	if err != nil {
		return model.ExercisePage{}, err
	}
	return newExercisePage(items, total, query), nil
}

func (s *ExerciseService) GetByID(ctx context.Context, id int64) (model.Exercise, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *ExerciseService) GetByMuscleGroup(
	ctx context.Context,
	muscleGroupID int64,
	query model.ExerciseQuery,
) (model.ExercisePage, error) {
	if _, err := s.muscleGroups.GetByID(ctx, muscleGroupID); err != nil {
		return model.ExercisePage{}, err
	}
	query = normalizeExerciseQuery(query)
	items, total, err := s.repository.GetByMuscleGroup(ctx, muscleGroupID, query)
	if err != nil {
		return model.ExercisePage{}, err
	}
	return newExercisePage(items, total, query), nil
}

func (s *ExerciseService) Create(
	ctx context.Context,
	name string,
	muscleGroupID int64,
) (model.Exercise, error) {
	name, err := validExerciseName(name)
	if err != nil {
		return model.Exercise{}, err
	}
	if _, err := s.muscleGroups.GetByID(ctx, muscleGroupID); err != nil {
		return model.Exercise{}, err
	}
	exercise, err := s.repository.Create(ctx, name, muscleGroupID)
	if errors.Is(err, repository.ErrExerciseMuscleGroupAbsent) {
		return model.Exercise{}, ErrMuscleGroupNotFound
	}
	return exercise, err
}

func (s *ExerciseService) Update(
	ctx context.Context,
	id int64,
	name *string,
	muscleGroupID *int64,
) (model.Exercise, error) {
	if name == nil && muscleGroupID == nil {
		return model.Exercise{}, ErrEmptyExerciseUpdate
	}
	if name != nil {
		normalizedName, err := validExerciseName(*name)
		if err != nil {
			return model.Exercise{}, err
		}
		name = &normalizedName
	}
	if muscleGroupID != nil {
		if *muscleGroupID < 1 {
			return model.Exercise{}, ErrInvalidMuscleGroupID
		}
		if _, err := s.muscleGroups.GetByID(ctx, *muscleGroupID); err != nil {
			return model.Exercise{}, err
		}
	}
	exercise, err := s.repository.Update(ctx, id, name, muscleGroupID)
	if errors.Is(err, repository.ErrExerciseMuscleGroupAbsent) {
		return model.Exercise{}, ErrMuscleGroupNotFound
	}
	return exercise, err
}

func (s *ExerciseService) Delete(ctx context.Context, id int64) error {
	deleted, err := s.repository.SoftDelete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrExerciseNotFound
	}
	return nil
}

func normalizeExerciseQuery(query model.ExerciseQuery) model.ExerciseQuery {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > MaxExercisePageSize {
		query.PageSize = DefaultExercisePageSize
	}
	query.Search = strings.TrimSpace(query.Search)
	if query.SortBy == "" {
		query.SortBy = "exercise_id"
	}
	if query.SortOrder == "" {
		query.SortOrder = "asc"
	}
	return query
}

func newExercisePage(
	items []model.Exercise,
	total int64,
	query model.ExerciseQuery,
) model.ExercisePage {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	}
	return model.ExercisePage{
		Items:      items,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}
}

func validExerciseName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 100 {
		return "", ErrInvalidExerciseName
	}
	return name, nil
}
