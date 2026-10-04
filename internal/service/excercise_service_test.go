package service

import (
	"context"
	"errors"
	"testing"

	"github.com/niviten/heavy/internal/model"
)

type stubExcerciseRepository struct {
	items       []model.Excercise
	total       int64
	excercise   model.Excercise
	err         error
	deleted     bool
	query       model.ExcerciseQuery
	muscleID    int64
	createdName string
	updatedName *string
}

func (r *stubExcerciseRepository) GetAll(
	_ context.Context,
	query model.ExcerciseQuery,
) ([]model.Excercise, int64, error) {
	r.query = query
	return r.items, r.total, r.err
}

func (r *stubExcerciseRepository) GetByMuscleGroup(
	_ context.Context,
	muscleGroupID int64,
	query model.ExcerciseQuery,
) ([]model.Excercise, int64, error) {
	r.muscleID = muscleGroupID
	r.query = query
	return r.items, r.total, r.err
}

func (r *stubExcerciseRepository) GetByID(context.Context, int64) (model.Excercise, error) {
	return r.excercise, r.err
}

func (r *stubExcerciseRepository) Create(
	_ context.Context,
	name string,
	muscleGroupID int64,
) (model.Excercise, error) {
	r.createdName = name
	r.muscleID = muscleGroupID
	return r.excercise, r.err
}

func (r *stubExcerciseRepository) Update(
	_ context.Context,
	_ int64,
	name *string,
	muscleGroupID *int64,
) (model.Excercise, error) {
	r.updatedName = name
	if muscleGroupID != nil {
		r.muscleID = *muscleGroupID
	}
	return r.excercise, r.err
}

func (r *stubExcerciseRepository) SoftDelete(context.Context, int64) (bool, error) {
	return r.deleted, r.err
}

type stubMuscleGroupLookup struct {
	err error
}

func (s stubMuscleGroupLookup) GetByID(context.Context, int64) (model.MuscleGroup, error) {
	return model.MuscleGroup{ID: 1}, s.err
}

func TestExcerciseServiceGetAllNormalizesAndPaginates(t *testing.T) {
	repository := &stubExcerciseRepository{
		items: []model.Excercise{{ID: 1, Name: "Bench Press", MuscleGroupID: 1}},
		total: 41,
	}
	service := NewExcerciseService(repository, stubMuscleGroupLookup{})

	page, err := service.GetAll(context.Background(), model.ExcerciseQuery{})
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if page.Page != 1 || page.PageSize != 20 || page.TotalPages != 3 || page.TotalItems != 41 {
		t.Errorf("page = %+v", page)
	}
	if repository.query.SortBy != "excercise_id" || repository.query.SortOrder != "asc" {
		t.Errorf("normalized query = %+v", repository.query)
	}
}

func TestExcerciseServiceCreateTrimsName(t *testing.T) {
	repository := &stubExcerciseRepository{
		excercise: model.Excercise{ID: 1, Name: "Bench Press", MuscleGroupID: 1},
	}
	service := NewExcerciseService(repository, stubMuscleGroupLookup{})

	_, err := service.Create(context.Background(), "  Bench Press  ", 1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repository.createdName != "Bench Press" || repository.muscleID != 1 {
		t.Errorf("created name = %q, muscle group = %d", repository.createdName, repository.muscleID)
	}
}

func TestExcerciseServiceRejectsInvalidName(t *testing.T) {
	service := NewExcerciseService(&stubExcerciseRepository{}, stubMuscleGroupLookup{})
	invalidName := "   "

	if _, err := service.Create(context.Background(), "   ", 1); !errors.Is(err, ErrInvalidExcerciseName) {
		t.Errorf("Create() error = %v, want ErrInvalidExcerciseName", err)
	}
	if _, err := service.Update(context.Background(), 1, &invalidName, nil); !errors.Is(err, ErrInvalidExcerciseName) {
		t.Errorf("Update() error = %v, want ErrInvalidExcerciseName", err)
	}
}

func TestExcerciseServiceUpdatesNameAndMuscleGroup(t *testing.T) {
	repository := &stubExcerciseRepository{excercise: model.Excercise{
		ID: 1, Name: "Incline Press", MuscleGroupID: 2,
	}}
	service := NewExcerciseService(repository, stubMuscleGroupLookup{})
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

func TestExcerciseServiceRejectsEmptyUpdate(t *testing.T) {
	service := NewExcerciseService(&stubExcerciseRepository{}, stubMuscleGroupLookup{})

	if _, err := service.Update(context.Background(), 1, nil, nil); !errors.Is(err, ErrEmptyExcerciseUpdate) {
		t.Errorf("Update() error = %v, want ErrEmptyExcerciseUpdate", err)
	}
}

func TestExcerciseServiceRejectsInvalidUpdateMuscleGroup(t *testing.T) {
	service := NewExcerciseService(
		&stubExcerciseRepository{},
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

func TestExcerciseServiceValidatesMuscleGroup(t *testing.T) {
	service := NewExcerciseService(
		&stubExcerciseRepository{},
		stubMuscleGroupLookup{err: ErrMuscleGroupNotFound},
	)

	if _, err := service.Create(context.Background(), "Bench Press", 99); !errors.Is(err, ErrMuscleGroupNotFound) {
		t.Errorf("Create() error = %v, want ErrMuscleGroupNotFound", err)
	}
	if _, err := service.GetByMuscleGroup(
		context.Background(),
		99,
		model.ExcerciseQuery{},
	); !errors.Is(err, ErrMuscleGroupNotFound) {
		t.Errorf("GetByMuscleGroup() error = %v, want ErrMuscleGroupNotFound", err)
	}
}

func TestExcerciseServiceDeleteNotFound(t *testing.T) {
	service := NewExcerciseService(&stubExcerciseRepository{}, stubMuscleGroupLookup{})

	if err := service.Delete(context.Background(), 99); !errors.Is(err, ErrExcerciseNotFound) {
		t.Errorf("Delete() error = %v, want ErrExcerciseNotFound", err)
	}
}
