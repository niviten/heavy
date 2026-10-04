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

type excerciseManager struct{}

func (excerciseManager) GetAll(context.Context, model.ExcerciseQuery) (model.ExcercisePage, error) {
	return model.ExcercisePage{Items: []model.Excercise{}, Page: 1, PageSize: 20}, nil
}

func (excerciseManager) GetByID(context.Context, int64) (model.Excercise, error) {
	return model.Excercise{ID: 1, Name: "Bench Press", MuscleGroupID: 1}, nil
}

func (excerciseManager) GetByMuscleGroup(
	context.Context,
	int64,
	model.ExcerciseQuery,
) (model.ExcercisePage, error) {
	return model.ExcercisePage{Items: []model.Excercise{}, Page: 1, PageSize: 20}, nil
}

func (excerciseManager) Create(context.Context, string, int64) (model.Excercise, error) {
	return model.Excercise{ID: 1, Name: "Bench Press", MuscleGroupID: 1}, nil
}

func (excerciseManager) Update(context.Context, int64, *string, *int64) (model.Excercise, error) {
	return model.Excercise{ID: 1, Name: "Bench Press", MuscleGroupID: 1}, nil
}

func (excerciseManager) Delete(context.Context, int64) error { return nil }

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

func TestRegister(t *testing.T) {
	e := echo.New()
	Register(
		e,
		controller.NewHealthController(healthyChecker{}),
		controller.NewMuscleGroupController(muscleGroupReader{}),
		controller.NewExcerciseController(excerciseManager{}),
		controller.NewWorkoutRoutineController(workoutRoutineManager{}),
	)

	for _, path := range []string{
		"/api/json",
		"/api/json/health",
		"/api/json/muscle-groups",
		"/api/json/muscle-groups/1",
		"/api/json/muscle-groups/name/Chest",
		"/api/json/excercises",
		"/api/json/excercises/1",
		"/api/json/muscle-groups/1/excercises",
		"/api/json/workout-routines",
		"/api/json/workout-routines/1",
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
}
