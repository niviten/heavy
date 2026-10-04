package service

import (
	"context"
	"errors"
	"testing"

	"github.com/niviten/heavy/internal/model"
)

type stubExerciseRepository struct {
	items       []model.Exercise
	total       int64
	exercise    model.Exercise
	err         error
	deleted     bool
	query       model.ExerciseQuery
	muscleID    int64
	createdName string
	updatedName *string
}

func (r *stubExerciseRepository) GetAll(
	_ context.Context,
	query model.ExerciseQuery,
) ([]model.Exercise, int64, error) {
	r.query = query
	return r.items, r.total, r.err
}

func (r *stubExerciseRepository) GetByMuscleGroup(
	_ context.Context,
	muscleGroupID int64,
	query model.ExerciseQuery,
) ([]model.Exercise, int64, error) {
	r.muscleID = muscleGroupID
	r.query = query
	return r.items, r.total, r.err
}

func (r *stubExerciseRepository) GetByID(context.Context, int64) (model.Exercise, error) {
	return r.exercise, r.err
}

func (r *stubExerciseRepository) Create(
	_ context.Context,
	name string,
	muscleGroupID int64,
) (model.Exercise, error) {
	r.createdName = name
	r.muscleID = muscleGroupID
	return r.exercise, r.err
}

func (r *stubExerciseRepository) Update(
	_ context.Context,
	_ int64,
	name *string,
	muscleGroupID *int64,
) (model.Exercise, error) {
	r.updatedName = name
	if muscleGroupID != nil {
		r.muscleID = *muscleGroupID
	}
	return r.exercise, r.err
}

func (r *stubExerciseRepository) SoftDelete(context.Context, int64) (bool, error) {
	return r.deleted, r.err
}

type stubMuscleGroupLookup struct {
	err error
}

func (s stubMuscleGroupLookup) GetByID(context.Context, int64) (model.MuscleGroup, error) {
	return model.MuscleGroup{ID: 1}, s.err
}

func TestExerciseServiceGetAllNormalizesAndPaginates(t *testing.T) {
	repository := &stubExerciseRepository{
		items: []model.Exercise{{ID: 1, Name: "Bench Press", MuscleGroupID: 1}},
		total: 41,
	}
	service := NewExerciseService(repository, stubMuscleGroupLookup{})

	page, err := service.GetAll(context.Background(), model.ExerciseQuery{})
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if page.Page != 1 || page.PageSize != 20 || page.TotalPages != 3 || page.TotalItems != 41 {
		t.Errorf("page = %+v", page)
	}
	if repository.query.SortBy != "exercise_id" || repository.query.SortOrder != "asc" {
		t.Errorf("normalized query = %+v", repository.query)
	}
}

func TestExerciseServiceCreateTrimsName(t *testing.T) {
	repository := &stubExerciseRepository{
		exercise: model.Exercise{ID: 1, Name: "Bench Press", MuscleGroupID: 1},
	}
	service := NewExerciseService(repository, stubMuscleGroupLookup{})

	_, err := service.Create(context.Background(), "  Bench Press  ", 1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repository.createdName != "Bench Press" || repository.muscleID != 1 {
		t.Errorf("created name = %q, muscle group = %d", repository.createdName, repository.muscleID)
	}
}

func TestExerciseServiceRejectsInvalidName(t *testing.T) {
	service := NewExerciseService(&stubExerciseRepository{}, stubMuscleGroupLookup{})
	invalidName := "   "

	if _, err := service.Create(context.Background(), "   ", 1); !errors.Is(err, ErrInvalidExerciseName) {
		t.Errorf("Create() error = %v, want ErrInvalidExerciseName", err)
	}
	if _, err := service.Update(context.Background(), 1, &invalidName, nil); !errors.Is(err, ErrInvalidExerciseName) {
		t.Errorf("Update() error = %v, want ErrInvalidExerciseName", err)
	}
}

func TestExerciseServiceUpdatesNameAndMuscleGroup(t *testing.T) {
	repository := &stubExerciseRepository{exercise: model.Exercise{
		ID: 1, Name: "Incline Press", MuscleGroupID: 2,
	}}
	service := NewExerciseService(repository, stubMuscleGroupLookup{})
	name := "  Incline Press  "
	muscleGroupID := int64(2)

	_, err := service.Update(context.Background(), 1, &name, &muscleGroupID)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if repository.updatedName == nil || *repository.updatedName != "Incline Press" {
		t.Errorf("updated name = %v, want Incline Press", repository.updatedName)
	}
	if repository.muscleID != 2 {
		t.Errorf("muscle group = %d, want 2", repository.muscleID)
	}
}

func TestExerciseServiceRejectsEmptyUpdate(t *testing.T) {
	service := NewExerciseService(&stubExerciseRepository{}, stubMuscleGroupLookup{})

	if _, err := service.Update(context.Background(), 1, nil, nil); !errors.Is(err, ErrEmptyExerciseUpdate) {
		t.Errorf("Update() error = %v, want ErrEmptyExerciseUpdate", err)
	}
}

func TestExerciseServiceRejectsInvalidUpdateMuscleGroup(t *testing.T) {
	service := NewExerciseService(
		&stubExerciseRepository{},
		stubMuscleGroupLookup{err: ErrMuscleGroupNotFound},
	)
	muscleGroupID := int64(99)

	if _, err := service.Update(
		context.Background(),
		1,
		nil,
		&muscleGroupID,
	); !errors.Is(err, ErrMuscleGroupNotFound) {
		t.Errorf("Update() error = %v, want ErrMuscleGroupNotFound", err)
	}
}

func TestExerciseServiceValidatesMuscleGroup(t *testing.T) {
	service := NewExerciseService(
		&stubExerciseRepository{},
		stubMuscleGroupLookup{err: ErrMuscleGroupNotFound},
	)

	if _, err := service.Create(context.Background(), "Bench Press", 99); !errors.Is(err, ErrMuscleGroupNotFound) {
		t.Errorf("Create() error = %v, want ErrMuscleGroupNotFound", err)
	}
	if _, err := service.GetByMuscleGroup(
		context.Background(),
		99,
		model.ExerciseQuery{},
	); !errors.Is(err, ErrMuscleGroupNotFound) {
		t.Errorf("GetByMuscleGroup() error = %v, want ErrMuscleGroupNotFound", err)
	}
}

func TestExerciseServiceDeleteNotFound(t *testing.T) {
	service := NewExerciseService(&stubExerciseRepository{}, stubMuscleGroupLookup{})

	if err := service.Delete(context.Background(), 99); !errors.Is(err, ErrExerciseNotFound) {
		t.Errorf("Delete() error = %v, want ErrExerciseNotFound", err)
	}
}
