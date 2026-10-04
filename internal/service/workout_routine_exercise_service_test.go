package service

import (
	"context"
	"errors"
	"testing"

	"github.com/niviten/heavy/internal/model"
)

type stubWorkoutRoutineExerciseRepository struct {
	items       []model.WorkoutRoutineExercise
	total       int64
	item        model.WorkoutRoutineExercise
	query       model.WorkoutRoutineExerciseQuery
	routineID   int64
	exerciseID  int64
	exerciseIDs []int64
	destination int
	deleted     bool
	err         error
}

func (s *stubWorkoutRoutineExerciseRepository) GetAll(
	_ context.Context,
	routineID int64,
	query model.WorkoutRoutineExerciseQuery,
) ([]model.WorkoutRoutineExercise, int64, error) {
	s.routineID = routineID
	s.query = query
	return s.items, s.total, s.err
}

func (s *stubWorkoutRoutineExerciseRepository) Create(
	_ context.Context,
	routineID int64,
	exerciseID int64,
) (model.WorkoutRoutineExercise, error) {
	s.routineID = routineID
	s.exerciseID = exerciseID
	return s.item, s.err
}

func (s *stubWorkoutRoutineExerciseRepository) CreateMany(
	_ context.Context,
	routineID int64,
	exerciseIDs []int64,
) ([]model.WorkoutRoutineExercise, error) {
	s.routineID = routineID
	s.exerciseIDs = exerciseIDs
	return s.items, s.err
}

func (s *stubWorkoutRoutineExerciseRepository) Delete(
	_ context.Context,
	routineID int64,
	exerciseID int64,
) (bool, error) {
	s.routineID = routineID
	s.exerciseID = exerciseID
	return s.deleted, s.err
}

func (s *stubWorkoutRoutineExerciseRepository) Reorder(
	_ context.Context,
	routineID int64,
	exerciseID int64,
	destination int,
) (model.WorkoutRoutineExercise, error) {
	s.routineID = routineID
	s.exerciseID = exerciseID
	s.destination = destination
	return s.item, s.err
}

func TestWorkoutRoutineExerciseServiceGetAllNormalizesAndPaginates(t *testing.T) {
	repository := &stubWorkoutRoutineExerciseRepository{
		items: []model.WorkoutRoutineExercise{{ExerciseID: 3}},
		total: 41,
	}
	exercises := NewWorkoutRoutineExerciseService(repository)

	page, err := exercises.GetAll(context.Background(), 7, model.WorkoutRoutineExerciseQuery{})
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if page.Page != 1 || page.PageSize != 20 || page.TotalItems != 41 || page.TotalPages != 3 {
		t.Errorf("page = %+v", page)
	}
	if repository.routineID != 7 || repository.query.SortBy != "exercise_order" ||
		repository.query.SortOrder != "asc" {
		t.Errorf("repository query = %+v, routine id = %d", repository.query, repository.routineID)
	}
}

func TestWorkoutRoutineExerciseServiceDeleteNotFound(t *testing.T) {
	exercises := NewWorkoutRoutineExerciseService(&stubWorkoutRoutineExerciseRepository{})
	if err := exercises.Delete(context.Background(), 1, 2); !errors.Is(
		err,
		ErrWorkoutRoutineExerciseNotFound,
	) {
		t.Fatalf("Delete() error = %v, want membership not found", err)
	}
}

func TestWorkoutRoutineExerciseServiceCreateManyValidatesInput(t *testing.T) {
	tests := []struct {
		name string
		ids  []int64
		err  error
	}{
		{name: "empty", ids: nil, err: ErrEmptyExerciseIDs},
		{name: "non-positive", ids: []int64{1, 0}, err: ErrInvalidExerciseID},
		{name: "duplicate", ids: []int64{1, 2, 1}, err: ErrDuplicateExerciseID},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &stubWorkoutRoutineExerciseRepository{}
			exercises := NewWorkoutRoutineExerciseService(repository)
			if _, err := exercises.CreateMany(context.Background(), 4, test.ids); !errors.Is(
				err,
				test.err,
			) {
				t.Fatalf("CreateMany() error = %v, want %v", err, test.err)
			}
			if repository.exerciseIDs != nil {
				t.Errorf("repository called with %v", repository.exerciseIDs)
			}
		})
	}
}

func TestWorkoutRoutineExerciseServiceCreateManyPreservesInputOrder(t *testing.T) {
	repository := &stubWorkoutRoutineExerciseRepository{}
	exercises := NewWorkoutRoutineExerciseService(repository)
	if _, err := exercises.CreateMany(context.Background(), 4, []int64{9, 3, 7}); err != nil {
		t.Fatalf("CreateMany() error = %v", err)
	}
	if repository.routineID != 4 || len(repository.exerciseIDs) != 3 ||
		repository.exerciseIDs[0] != 9 || repository.exerciseIDs[1] != 3 ||
		repository.exerciseIDs[2] != 7 {
		t.Errorf("routine id = %d, exercise ids = %v", repository.routineID, repository.exerciseIDs)
	}
}

func TestWorkoutRoutineExerciseServiceReorderValidatesAndDelegates(t *testing.T) {
	repository := &stubWorkoutRoutineExerciseRepository{}
	exercises := NewWorkoutRoutineExerciseService(repository)

	if _, err := exercises.Reorder(context.Background(), 1, 2, 0); !errors.Is(
		err,
		ErrInvalidExercisePosition,
	) {
		t.Fatalf("Reorder() error = %v, want invalid position", err)
	}
	if repository.destination != 0 {
		t.Fatal("repository called for invalid destination")
	}
	if _, err := exercises.Reorder(context.Background(), 4, 9, 3); err != nil {
		t.Fatalf("Reorder() error = %v", err)
	}
	if repository.routineID != 4 || repository.exerciseID != 9 || repository.destination != 3 {
		t.Errorf("repository call = routine %d, exercise %d, destination %d",
			repository.routineID, repository.exerciseID, repository.destination)
	}
}
