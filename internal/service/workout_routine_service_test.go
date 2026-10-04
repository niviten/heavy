package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/niviten/heavy/internal/model"
)

type stubWorkoutRoutineRepository struct {
	items       []model.WorkoutRoutine
	total       int64
	routine     model.WorkoutRoutine
	err         error
	deleted     bool
	query       model.WorkoutRoutineQuery
	createdName string
	description *string
	updatedName string
	updatedDesc *string
	updateDesc  bool
}

func (r *stubWorkoutRoutineRepository) GetAll(
	_ context.Context,
	query model.WorkoutRoutineQuery,
) ([]model.WorkoutRoutine, int64, error) {
	r.query = query
	return r.items, r.total, r.err
}

func (r *stubWorkoutRoutineRepository) GetByID(
	context.Context,
	int64,
) (model.WorkoutRoutine, error) {
	return r.routine, r.err
}

func (r *stubWorkoutRoutineRepository) Create(
	_ context.Context,
	name string,
	description *string,
) (model.WorkoutRoutine, error) {
	r.createdName = name
	r.description = description
	return r.routine, r.err
}

func (r *stubWorkoutRoutineRepository) Update(
	_ context.Context,
	_ int64,
	name *string,
	description *string,
	updateDescription bool,
) (model.WorkoutRoutine, error) {
	if name != nil {
		r.updatedName = *name
	}
	r.updatedDesc = description
	r.updateDesc = updateDescription
	return r.routine, r.err
}

func (r *stubWorkoutRoutineRepository) Delete(context.Context, int64) (bool, error) {
	return r.deleted, r.err
}

func TestWorkoutRoutineServiceGetAllNormalizesAndPaginates(t *testing.T) {
	repository := &stubWorkoutRoutineRepository{
		items: []model.WorkoutRoutine{{ID: 1, Name: "Push Day"}},
		total: 41,
	}
	workoutRoutines := NewWorkoutRoutineService(repository)

	page, err := workoutRoutines.GetAll(context.Background(), model.WorkoutRoutineQuery{})
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if page.Page != 1 || page.PageSize != 20 || page.TotalPages != 3 || page.TotalItems != 41 {
		t.Errorf("page = %+v", page)
	}
	if repository.query.SortBy != "workout_routine_id" || repository.query.SortOrder != "asc" {
		t.Errorf("normalized query = %+v", repository.query)
	}
}

func TestWorkoutRoutineServiceCreateNormalizesFields(t *testing.T) {
	description := "  Chest and shoulders  "
	repository := &stubWorkoutRoutineRepository{routine: model.WorkoutRoutine{ID: 1, Name: "Push Day"}}
	workoutRoutines := NewWorkoutRoutineService(repository)

	_, err := workoutRoutines.Create(context.Background(), "  Push Day  ", &description)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repository.createdName != "Push Day" {
		t.Errorf("created name = %q", repository.createdName)
	}
	if repository.description == nil || *repository.description != "Chest and shoulders" {
		t.Errorf("description = %v", repository.description)
	}
}

func TestWorkoutRoutineServiceValidatesFields(t *testing.T) {
	workoutRoutines := NewWorkoutRoutineService(&stubWorkoutRoutineRepository{})
	longDescription := strings.Repeat("x", 256)

	if _, err := workoutRoutines.Create(context.Background(), "  ", nil); !errors.Is(err, ErrInvalidWorkoutRoutineName) {
		t.Errorf("Create() error = %v, want ErrInvalidWorkoutRoutineName", err)
	}
	if _, err := workoutRoutines.Create(context.Background(), "Push Day", &longDescription); !errors.Is(err, ErrInvalidWorkoutRoutineDescription) {
		t.Errorf("Create() error = %v, want ErrInvalidWorkoutRoutineDescription", err)
	}
	if _, err := workoutRoutines.Update(context.Background(), 1, nil, nil); !errors.Is(err, ErrEmptyWorkoutRoutineUpdate) {
		t.Errorf("Update() error = %v, want ErrEmptyWorkoutRoutineUpdate", err)
	}
}

func TestWorkoutRoutineServiceUpdateTrimsName(t *testing.T) {
	repository := &stubWorkoutRoutineRepository{}
	workoutRoutines := NewWorkoutRoutineService(repository)
	name := "  Upper Push  "

	if _, err := workoutRoutines.Update(context.Background(), 1, &name, nil); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if repository.updatedName != "Upper Push" {
		t.Errorf("updated name = %q", repository.updatedName)
	}
}

func TestWorkoutRoutineServiceUpdatesDescription(t *testing.T) {
	repository := &stubWorkoutRoutineRepository{}
	workoutRoutines := NewWorkoutRoutineService(repository)
	description := "  Chest, shoulders, and triceps  "

	if _, err := workoutRoutines.Update(context.Background(), 1, nil, &description); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !repository.updateDesc || repository.updatedDesc == nil || *repository.updatedDesc != "Chest, shoulders, and triceps" {
		t.Errorf("updated description = %v", repository.updatedDesc)
	}
}

func TestWorkoutRoutineServiceConvertsBlankDescriptionToNull(t *testing.T) {
	repository := &stubWorkoutRoutineRepository{}
	workoutRoutines := NewWorkoutRoutineService(repository)
	description := "   "

	if _, err := workoutRoutines.Update(context.Background(), 1, nil, &description); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !repository.updateDesc {
		t.Fatal("description update was not requested")
	}
	if repository.updatedDesc != nil {
		t.Errorf("updated description = %q, want nil", *repository.updatedDesc)
	}
}

func TestWorkoutRoutineServiceDeleteNotFound(t *testing.T) {
	workoutRoutines := NewWorkoutRoutineService(&stubWorkoutRoutineRepository{})

	if err := workoutRoutines.Delete(context.Background(), 99); !errors.Is(err, ErrWorkoutRoutineNotFound) {
		t.Errorf("Delete() error = %v, want ErrWorkoutRoutineNotFound", err)
	}
}
