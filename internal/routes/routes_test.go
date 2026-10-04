package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/niviten/heavy/internal/controller"
	"github.com/niviten/heavy/internal/dto"
	"github.com/niviten/heavy/internal/model"
)

type healthyChecker struct{}

func (healthyChecker) Check(context.Context) (dto.HealthResponse, error) {
	return dto.HealthResponse{Status: "ok", Database: "up"}, nil
}

type muscleGroupReader struct{}

func (muscleGroupReader) GetAll(context.Context) ([]model.MuscleGroup, error) {
	return []model.MuscleGroup{{ID: 1, Name: "chest", DisplayName: "Chest"}}, nil
}

func (muscleGroupReader) GetByID(context.Context, int64) (model.MuscleGroup, error) {
	return model.MuscleGroup{ID: 1, Name: "chest", DisplayName: "Chest"}, nil
}

func (muscleGroupReader) GetByName(context.Context, string) (model.MuscleGroup, error) {
	return model.MuscleGroup{ID: 1, Name: "chest", DisplayName: "Chest"}, nil
}

type exerciseManager struct{}

func (exerciseManager) GetAll(context.Context, model.ExerciseQuery) (model.ExercisePage, error) {
	return model.ExercisePage{Items: []model.Exercise{}, Page: 1, PageSize: 20}, nil
}

func (exerciseManager) GetByID(context.Context, int64) (model.Exercise, error) {
	return model.Exercise{ID: 1, Name: "Bench Press", MuscleGroupID: 1}, nil
}

func (exerciseManager) GetByMuscleGroup(
	context.Context,
	int64,
	model.ExerciseQuery,
) (model.ExercisePage, error) {
	return model.ExercisePage{Items: []model.Exercise{}, Page: 1, PageSize: 20}, nil
}

func (exerciseManager) Create(context.Context, string, int64) (model.Exercise, error) {
	return model.Exercise{ID: 1, Name: "Bench Press", MuscleGroupID: 1}, nil
}

func (exerciseManager) Update(context.Context, int64, *string, *int64) (model.Exercise, error) {
	return model.Exercise{ID: 1, Name: "Bench Press", MuscleGroupID: 1}, nil
}

func (exerciseManager) Delete(context.Context, int64) error { return nil }

type workoutRoutineManager struct{}

func (workoutRoutineManager) GetAll(
	context.Context,
	model.WorkoutRoutineQuery,
) (model.WorkoutRoutinePage, error) {
	return model.WorkoutRoutinePage{Items: []model.WorkoutRoutine{}, Page: 1, PageSize: 20}, nil
}

func (workoutRoutineManager) GetByID(context.Context, int64) (model.WorkoutRoutine, error) {
	return model.WorkoutRoutine{ID: 1, Name: "Push Day"}, nil
}

func (workoutRoutineManager) Create(
	context.Context,
	string,
	*string,
) (model.WorkoutRoutine, error) {
	return model.WorkoutRoutine{ID: 1, Name: "Push Day"}, nil
}

func (workoutRoutineManager) Update(
	context.Context,
	int64,
	*string,
	*string,
) (model.WorkoutRoutine, error) {
	return model.WorkoutRoutine{ID: 1, Name: "Push Day"}, nil
}

func (workoutRoutineManager) Delete(context.Context, int64) error { return nil }

type workoutRoutineExerciseManager struct{}

func (workoutRoutineExerciseManager) GetAll(
	context.Context,
	int64,
	model.WorkoutRoutineExerciseQuery,
) (model.WorkoutRoutineExercisePage, error) {
	return model.WorkoutRoutineExercisePage{
		Items: []model.WorkoutRoutineExercise{}, Page: 1, PageSize: 20,
	}, nil
}

func (workoutRoutineExerciseManager) Create(
	context.Context,
	int64,
	int64,
) (model.WorkoutRoutineExercise, error) {
	return model.WorkoutRoutineExercise{}, nil
}

func (workoutRoutineExerciseManager) CreateMany(
	context.Context,
	int64,
	[]int64,
) ([]model.WorkoutRoutineExercise, error) {
	return []model.WorkoutRoutineExercise{}, nil
}

func (workoutRoutineExerciseManager) Delete(context.Context, int64, int64) error { return nil }

func (workoutRoutineExerciseManager) Reorder(
	context.Context,
	int64,
	int64,
	int,
) (model.WorkoutRoutineExercise, error) {
	return model.WorkoutRoutineExercise{}, nil
}

func TestRegister(t *testing.T) {
	e := echo.New()
	Register(
		e,
		controller.NewHealthController(healthyChecker{}),
		controller.NewMuscleGroupController(muscleGroupReader{}),
		controller.NewExerciseController(exerciseManager{}),
		controller.NewWorkoutRoutineController(workoutRoutineManager{}),
		controller.NewWorkoutRoutineExerciseController(workoutRoutineExerciseManager{}),
	)

	for _, path := range []string{
		"/api/json",
		"/api/json/health",
		"/api/json/muscle-groups",
		"/api/json/muscle-groups/1",
		"/api/json/muscle-groups/name/Chest",
		"/api/json/exercises",
		"/api/json/exercises/1",
		"/api/json/muscle-groups/1/exercises",
		"/api/json/workout-routines",
		"/api/json/workout-routines/1",
		"/api/json/workout-routines/1/exercises",
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		e.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}
	}

	for _, path := range []string{"/", "/health", "/muscle-groups"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		e.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Errorf("GET %s status = %d, want %d", path, recorder.Code, http.StatusNotFound)
		}
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/json/workout-routines/1/exercises/bulk",
		nil,
	)
	e.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Errorf("POST bulk routine exercises status = %d, want %d",
			recorder.Code, http.StatusCreated)
	}
}
