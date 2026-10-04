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
	DefaultWorkoutRoutinePageSize = 20
	MaxWorkoutRoutinePageSize     = 100
)

var (
	ErrWorkoutRoutineNotFound           = repository.ErrWorkoutRoutineNotFound
	ErrWorkoutRoutineNameConflict       = repository.ErrWorkoutRoutineNameConflict
	ErrInvalidWorkoutRoutineName        = errors.New("workout routine name must contain between 1 and 100 characters")
	ErrInvalidWorkoutRoutineDescription = errors.New("workout routine description cannot exceed 255 characters")
	ErrEmptyWorkoutRoutineUpdate        = errors.New("at least one workout routine field must be provided")
)

// WorkoutRoutineRepository is the persistence behavior required by the service.
type WorkoutRoutineRepository interface {
	GetAll(context.Context, model.WorkoutRoutineQuery) ([]model.WorkoutRoutine, int64, error)
	GetByID(context.Context, int64) (model.WorkoutRoutine, error)
	Create(context.Context, string, *string) (model.WorkoutRoutine, error)
	Update(context.Context, int64, *string, *string, bool) (model.WorkoutRoutine, error)
	Delete(context.Context, int64) (bool, error)
}

type WorkoutRoutineService struct {
	repository WorkoutRoutineRepository
}

func NewWorkoutRoutineService(repository WorkoutRoutineRepository) *WorkoutRoutineService {
	return &WorkoutRoutineService{repository: repository}
}

func (s *WorkoutRoutineService) GetAll(
	ctx context.Context,
	query model.WorkoutRoutineQuery,
) (model.WorkoutRoutinePage, error) {
	query = normalizeWorkoutRoutineQuery(query)
	items, total, err := s.repository.GetAll(ctx, query)
	if err != nil {
		return model.WorkoutRoutinePage{}, err
	}
	return newWorkoutRoutinePage(items, total, query), nil
}

func (s *WorkoutRoutineService) GetByID(
	ctx context.Context,
	id int64,
) (model.WorkoutRoutine, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *WorkoutRoutineService) Create(
	ctx context.Context,
	name string,
	description *string,
) (model.WorkoutRoutine, error) {
	name, err := validWorkoutRoutineName(name)
	if err != nil {
		return model.WorkoutRoutine{}, err
	}
	description, err = validWorkoutRoutineDescription(description)
	if err != nil {
		return model.WorkoutRoutine{}, err
	}
	return s.repository.Create(ctx, name, description)
}

func (s *WorkoutRoutineService) Update(
	ctx context.Context,
	id int64,
	name *string,
	description *string,
) (model.WorkoutRoutine, error) {
	if name == nil && description == nil {
		return model.WorkoutRoutine{}, ErrEmptyWorkoutRoutineUpdate
	}
	updateDescription := description != nil
	if name != nil {
		normalizedName, err := validWorkoutRoutineName(*name)
		if err != nil {
			return model.WorkoutRoutine{}, err
		}
		name = &normalizedName
	}
	if description != nil {
		normalizedDescription := strings.TrimSpace(*description)
		if utf8.RuneCountInString(normalizedDescription) > 255 {
			return model.WorkoutRoutine{}, ErrInvalidWorkoutRoutineDescription
		}
		if normalizedDescription == "" {
			description = nil
		} else {
			description = &normalizedDescription
		}
	}
	return s.repository.Update(ctx, id, name, description, updateDescription)
}

func (s *WorkoutRoutineService) Delete(ctx context.Context, id int64) error {
	deleted, err := s.repository.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrWorkoutRoutineNotFound
	}
	return nil
}

func normalizeWorkoutRoutineQuery(query model.WorkoutRoutineQuery) model.WorkoutRoutineQuery {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > MaxWorkoutRoutinePageSize {
		query.PageSize = DefaultWorkoutRoutinePageSize
	}
	query.Search = strings.TrimSpace(query.Search)
	if query.SortBy == "" {
		query.SortBy = "workout_routine_id"
	}
	if query.SortOrder == "" {
		query.SortOrder = "asc"
	}
	return query
}

func newWorkoutRoutinePage(
	items []model.WorkoutRoutine,
	total int64,
	query model.WorkoutRoutineQuery,
) model.WorkoutRoutinePage {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	}
	return model.WorkoutRoutinePage{
		Items:      items,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}
}

func validWorkoutRoutineName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 100 {
		return "", ErrInvalidWorkoutRoutineName
	}
	return name, nil
}

func validWorkoutRoutineDescription(description *string) (*string, error) {
	if description == nil {
		return nil, nil
	}
	normalized := strings.TrimSpace(*description)
	if utf8.RuneCountInString(normalized) > 255 {
		return nil, ErrInvalidWorkoutRoutineDescription
	}
	if normalized == "" {
		return nil, nil
	}
	return &normalized, nil
}
