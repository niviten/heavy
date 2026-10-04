package service

import (
	"context"
	"errors"
	"strings"

	"github.com/niviten/heavy/internal/model"
	"github.com/niviten/heavy/internal/repository"
)

const (
	DefaultWorkoutRoutineExercisePageSize = 20
	MaxWorkoutRoutineExercisePageSize     = 100
)

var (
	ErrWorkoutRoutineExerciseNotFound = repository.ErrWorkoutRoutineExerciseNotFound
	ErrWorkoutRoutineExerciseConflict = repository.ErrWorkoutRoutineExerciseConflict
	ErrInvalidExercisePosition        = repository.ErrInvalidExercisePosition
	ErrEmptyExerciseIDs               = errors.New("exercise ids must not be empty")
	ErrInvalidExerciseID              = errors.New("exercise ids must be positive integers")
	ErrDuplicateExerciseID            = errors.New("exercise ids must not contain duplicates")
)

// WorkoutRoutineExerciseRepository is the persistence behavior required by the service.
type WorkoutRoutineExerciseRepository interface {
	GetAll(
		context.Context,
		int64,
		model.WorkoutRoutineExerciseQuery,
	) ([]model.WorkoutRoutineExercise, int64, error)
	Create(context.Context, int64, int64) (model.WorkoutRoutineExercise, error)
	CreateMany(context.Context, int64, []int64) ([]model.WorkoutRoutineExercise, error)
	Delete(context.Context, int64, int64) (bool, error)
	Reorder(context.Context, int64, int64, int) (model.WorkoutRoutineExercise, error)
}

type WorkoutRoutineExerciseService struct {
	repository WorkoutRoutineExerciseRepository
}

func NewWorkoutRoutineExerciseService(
	repository WorkoutRoutineExerciseRepository,
) *WorkoutRoutineExerciseService {
	return &WorkoutRoutineExerciseService{repository: repository}
}

func (s *WorkoutRoutineExerciseService) GetAll(
	ctx context.Context,
	workoutRoutineID int64,
	query model.WorkoutRoutineExerciseQuery,
) (model.WorkoutRoutineExercisePage, error) {
	query = normalizeWorkoutRoutineExerciseQuery(query)
	items, total, err := s.repository.GetAll(ctx, workoutRoutineID, query)
	if err != nil {
		return model.WorkoutRoutineExercisePage{}, err
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	}
	return model.WorkoutRoutineExercisePage{
		Items:      items,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

func (s *WorkoutRoutineExerciseService) Create(
	ctx context.Context,
	workoutRoutineID int64,
	exerciseID int64,
) (model.WorkoutRoutineExercise, error) {
	return s.repository.Create(ctx, workoutRoutineID, exerciseID)
}

func (s *WorkoutRoutineExerciseService) CreateMany(
	ctx context.Context,
	workoutRoutineID int64,
	exerciseIDs []int64,
) ([]model.WorkoutRoutineExercise, error) {
	if len(exerciseIDs) == 0 {
		return nil, ErrEmptyExerciseIDs
	}
	seen := make(map[int64]struct{}, len(exerciseIDs))
	for _, exerciseID := range exerciseIDs {
		if exerciseID < 1 {
			return nil, ErrInvalidExerciseID
		}
		if _, exists := seen[exerciseID]; exists {
			return nil, ErrDuplicateExerciseID
		}
		seen[exerciseID] = struct{}{}
	}
	return s.repository.CreateMany(ctx, workoutRoutineID, exerciseIDs)
}

func (s *WorkoutRoutineExerciseService) Delete(
	ctx context.Context,
	workoutRoutineID int64,
	exerciseID int64,
) error {
	deleted, err := s.repository.Delete(ctx, workoutRoutineID, exerciseID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrWorkoutRoutineExerciseNotFound
	}
	return nil
}

func (s *WorkoutRoutineExerciseService) Reorder(
	ctx context.Context,
	workoutRoutineID int64,
	exerciseID int64,
	destinationPosition int,
) (model.WorkoutRoutineExercise, error) {
	if destinationPosition < 1 {
		return model.WorkoutRoutineExercise{}, ErrInvalidExercisePosition
	}
	return s.repository.Reorder(ctx, workoutRoutineID, exerciseID, destinationPosition)
}

func normalizeWorkoutRoutineExerciseQuery(
	query model.WorkoutRoutineExerciseQuery,
) model.WorkoutRoutineExerciseQuery {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > MaxWorkoutRoutineExercisePageSize {
		query.PageSize = DefaultWorkoutRoutineExercisePageSize
	}
	query.Search = strings.TrimSpace(query.Search)
	if query.SortBy == "" {
		query.SortBy = "exercise_order"
	}
	if query.SortOrder == "" {
		query.SortOrder = "asc"
	}
	return query
}
