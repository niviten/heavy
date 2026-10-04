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
	DefaultExcercisePageSize = 20
	MaxExcercisePageSize     = 100
)

var (
	ErrExcerciseNotFound     = repository.ErrExcerciseNotFound
	ErrExcerciseNameConflict = repository.ErrExcerciseNameConflict
	ErrInvalidExcerciseName  = errors.New("excercise name must contain between 1 and 100 characters")
	ErrInvalidMuscleGroupID  = errors.New("muscle group id must be a positive integer")
	ErrEmptyExcerciseUpdate  = errors.New("at least one excercise field must be provided")
)

// ExcerciseRepository is the persistence behavior required by the service.
type ExcerciseRepository interface {
	GetAll(context.Context, model.ExcerciseQuery) ([]model.Excercise, int64, error)
	GetByMuscleGroup(context.Context, int64, model.ExcerciseQuery) ([]model.Excercise, int64, error)
	GetByID(context.Context, int64) (model.Excercise, error)
	Create(context.Context, string, int64) (model.Excercise, error)
	Update(context.Context, int64, *string, *int64) (model.Excercise, error)
	SoftDelete(context.Context, int64) (bool, error)
}

// MuscleGroupLookup verifies referenced muscle groups.
type MuscleGroupLookup interface {
	GetByID(context.Context, int64) (model.MuscleGroup, error)
}

type ExcerciseService struct {
	repository   ExcerciseRepository
	muscleGroups MuscleGroupLookup
}

func NewExcerciseService(
	repository ExcerciseRepository,
	muscleGroups MuscleGroupLookup,
) *ExcerciseService {
	return &ExcerciseService{repository: repository, muscleGroups: muscleGroups}
}

func (s *ExcerciseService) GetAll(
	ctx context.Context,
	query model.ExcerciseQuery,
) (model.ExcercisePage, error) {
	query = normalizeExcerciseQuery(query)
	items, total, err := s.repository.GetAll(ctx, query)
	if err != nil {
		return model.ExcercisePage{}, err
	}
	return newExcercisePage(items, total, query), nil
}

func (s *ExcerciseService) GetByID(ctx context.Context, id int64) (model.Excercise, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *ExcerciseService) GetByMuscleGroup(
	ctx context.Context,
	muscleGroupID int64,
	query model.ExcerciseQuery,
) (model.ExcercisePage, error) {
	if _, err := s.muscleGroups.GetByID(ctx, muscleGroupID); err != nil {
		return model.ExcercisePage{}, err
	}
	query = normalizeExcerciseQuery(query)
	items, total, err := s.repository.GetByMuscleGroup(ctx, muscleGroupID, query)
	if err != nil {
		return model.ExcercisePage{}, err
	}
	return newExcercisePage(items, total, query), nil
}

func (s *ExcerciseService) Create(
	ctx context.Context,
	name string,
	muscleGroupID int64,
) (model.Excercise, error) {
	name, err := validExcerciseName(name)
	if err != nil {
		return model.Excercise{}, err
	}
	if _, err := s.muscleGroups.GetByID(ctx, muscleGroupID); err != nil {
		return model.Excercise{}, err
	}
	excercise, err := s.repository.Create(ctx, name, muscleGroupID)
	if errors.Is(err, repository.ErrExcerciseMuscleGroupAbsent) {
		return model.Excercise{}, ErrMuscleGroupNotFound
	}
	return excercise, err
}

func (s *ExcerciseService) Update(
	ctx context.Context,
	id int64,
	name *string,
	muscleGroupID *int64,
) (model.Excercise, error) {
	if name == nil && muscleGroupID == nil {
		return model.Excercise{}, ErrEmptyExcerciseUpdate
	}
	if name != nil {
		normalizedName, err := validExcerciseName(*name)
		if err != nil {
			return model.Excercise{}, err
		}
		name = &normalizedName
	}
	if muscleGroupID != nil {
		if *muscleGroupID < 1 {
			return model.Excercise{}, ErrInvalidMuscleGroupID
		}
		if _, err := s.muscleGroups.GetByID(ctx, *muscleGroupID); err != nil {
			return model.Excercise{}, err
		}
	}
	excercise, err := s.repository.Update(ctx, id, name, muscleGroupID)
	if errors.Is(err, repository.ErrExcerciseMuscleGroupAbsent) {
		return model.Excercise{}, ErrMuscleGroupNotFound
	}
	return excercise, err
}

func (s *ExcerciseService) Delete(ctx context.Context, id int64) error {
	deleted, err := s.repository.SoftDelete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrExcerciseNotFound
	}
	return nil
}

func normalizeExcerciseQuery(query model.ExcerciseQuery) model.ExcerciseQuery {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > MaxExcercisePageSize {
		query.PageSize = DefaultExcercisePageSize
	}
	query.Search = strings.TrimSpace(query.Search)
	if query.SortBy == "" {
		query.SortBy = "excercise_id"
	}
	if query.SortOrder == "" {
		query.SortOrder = "asc"
	}
	return query
}

func newExcercisePage(
	items []model.Excercise,
	total int64,
	query model.ExcerciseQuery,
) model.ExcercisePage {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	}
	return model.ExcercisePage{
		Items:      items,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}
}

func validExcerciseName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 100 {
		return "", ErrInvalidExcerciseName
	}
	return name, nil
}
