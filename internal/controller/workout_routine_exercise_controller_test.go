package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/niviten/heavy/internal/model"
	"github.com/niviten/heavy/internal/service"
)

type stubWorkoutRoutineExerciseManager struct {
	page        model.WorkoutRoutineExercisePage
	item        model.WorkoutRoutineExercise
	query       model.WorkoutRoutineExerciseQuery
	routineID   int64
	exerciseID  int64
	exerciseIDs []int64
	destination int
	deleteCall  bool
	err         error
}

func (s *stubWorkoutRoutineExerciseManager) GetAll(
	_ context.Context,
	routineID int64,
	query model.WorkoutRoutineExerciseQuery,
) (model.WorkoutRoutineExercisePage, error) {
	s.routineID = routineID
	s.query = query
	return s.page, s.err
}

func (s *stubWorkoutRoutineExerciseManager) Create(
	_ context.Context,
	routineID int64,
	exerciseID int64,
) (model.WorkoutRoutineExercise, error) {
	s.routineID = routineID
	s.exerciseID = exerciseID
	return s.item, s.err
}

func (s *stubWorkoutRoutineExerciseManager) CreateMany(
	_ context.Context,
	routineID int64,
	exerciseIDs []int64,
) ([]model.WorkoutRoutineExercise, error) {
	s.routineID = routineID
	s.exerciseIDs = exerciseIDs
	return []model.WorkoutRoutineExercise{s.item}, s.err
}

func (s *stubWorkoutRoutineExerciseManager) Delete(
	_ context.Context,
	routineID int64,
	exerciseID int64,
) error {
	s.routineID = routineID
	s.exerciseID = exerciseID
	s.deleteCall = true
	return s.err
}

func (s *stubWorkoutRoutineExerciseManager) Reorder(
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

func TestWorkoutRoutineExerciseControllerGetAll(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/json/workout-routines/4/exercises?page=2&page_size=5&search=press&sort_by=exercise_name&sort_order=desc",
		nil,
	)
	ctx := e.NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "routine_id", Value: "4"}})
	manager := &stubWorkoutRoutineExerciseManager{
		page: model.WorkoutRoutineExercisePage{
			Items: []model.WorkoutRoutineExercise{{
				ID:               8,
				WorkoutRoutineID: 4,
				ExerciseID:       12,
				ExerciseName:     "Bench Press",
				MuscleGroupID:    1,
				ExerciseOrder:    1500,
			}},
			Page: 2, PageSize: 5, TotalItems: 7, TotalPages: 2,
		},
	}
	controller := NewWorkoutRoutineExerciseController(manager)

	if err := controller.GetAll(ctx); err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	wantBody := "{\"data\":[{\"workout_routine_exercise_id\":8," +
		"\"workout_routine_id\":4,\"exercise_id\":12,\"exercise_name\":\"Bench Press\"," +
		"\"muscle_group_id\":1,\"exercise_order\":\"1.500\"}]," +
		"\"pagination\":{\"page\":2,\"page_size\":5,\"total_items\":7,\"total_pages\":2}}\n"
	if recorder.Body.String() != wantBody {
		t.Errorf("body = %q, want %q", recorder.Body.String(), wantBody)
	}
	if manager.routineID != 4 || manager.query.Page != 2 || manager.query.PageSize != 5 ||
		manager.query.Search != "press" || manager.query.SortBy != "exercise_name" ||
		manager.query.SortOrder != "desc" {
		t.Errorf("routine id = %d, query = %+v", manager.routineID, manager.query)
	}
}

func TestWorkoutRoutineExerciseControllerCreate(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/json/workout-routines/4/exercises",
		strings.NewReader(`{"exercise_id":12}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	ctx := e.NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "routine_id", Value: "4"}})
	manager := &stubWorkoutRoutineExerciseManager{item: model.WorkoutRoutineExercise{
		ID: 8, WorkoutRoutineID: 4, ExerciseID: 12, ExerciseName: "Bench Press",
		MuscleGroupID: 1, ExerciseOrder: 3000,
	}}
	controller := NewWorkoutRoutineExerciseController(manager)

	if err := controller.Create(ctx); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if recorder.Code != http.StatusCreated || manager.routineID != 4 || manager.exerciseID != 12 {
		t.Errorf("status = %d, routine id = %d, exercise id = %d",
			recorder.Code, manager.routineID, manager.exerciseID)
	}
}

func TestWorkoutRoutineExerciseControllerCreateMany(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/json/workout-routines/4/exercises/bulk",
		strings.NewReader(`{"exercise_ids":[12,15]}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	ctx := e.NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "routine_id", Value: "4"}})
	manager := &stubWorkoutRoutineExerciseManager{item: model.WorkoutRoutineExercise{
		ID: 8, WorkoutRoutineID: 4, ExerciseID: 12, ExerciseName: "Bench Press",
		MuscleGroupID: 1, ExerciseOrder: 3000,
	}}
	controller := NewWorkoutRoutineExerciseController(manager)

	if err := controller.CreateMany(ctx); err != nil {
		t.Fatalf("CreateMany() error = %v", err)
	}
	if recorder.Code != http.StatusCreated || manager.routineID != 4 ||
		len(manager.exerciseIDs) != 2 || manager.exerciseIDs[0] != 12 || manager.exerciseIDs[1] != 15 {
		t.Errorf("status = %d, routine id = %d, exercise ids = %v",
			recorder.Code, manager.routineID, manager.exerciseIDs)
	}
	if !strings.Contains(recorder.Body.String(), `"exercise_order":"3.000"`) {
		t.Errorf("body = %q", recorder.Body.String())
	}
}

func TestWorkoutRoutineExerciseControllerDelete(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/api/json/workout-routines/4/exercises/12", nil)
	ctx := e.NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{
		{Name: "routine_id", Value: "4"},
		{Name: "exercise_id", Value: "12"},
	})
	manager := &stubWorkoutRoutineExerciseManager{}
	controller := NewWorkoutRoutineExerciseController(manager)

	if err := controller.Delete(ctx); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if recorder.Code != http.StatusNoContent || !manager.deleteCall ||
		manager.routineID != 4 || manager.exerciseID != 12 {
		t.Errorf("status = %d, delete call = %v, routine id = %d, exercise id = %d",
			recorder.Code, manager.deleteCall, manager.routineID, manager.exerciseID)
	}
}

func TestWorkoutRoutineExerciseControllerReorder(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/json/workout-routines/4/exercises/12/order",
		strings.NewReader(`{"destination_position":2}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	ctx := e.NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{
		{Name: "routine_id", Value: "4"},
		{Name: "exercise_id", Value: "12"},
	})
	manager := &stubWorkoutRoutineExerciseManager{item: model.WorkoutRoutineExercise{
		ID: 8, WorkoutRoutineID: 4, ExerciseID: 12, ExerciseOrder: 1500,
	}}
	controller := NewWorkoutRoutineExerciseController(manager)

	if err := controller.Reorder(ctx); err != nil {
		t.Fatalf("Reorder() error = %v", err)
	}
	if recorder.Code != http.StatusOK || manager.routineID != 4 ||
		manager.exerciseID != 12 || manager.destination != 2 {
		t.Errorf("status = %d, routine id = %d, exercise id = %d, destination = %d",
			recorder.Code, manager.routineID, manager.exerciseID, manager.destination)
	}
}

func TestWorkoutRoutineExerciseControllerRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		call   func(*WorkoutRoutineExerciseController, *echo.Context) error
	}{
		{
			name: "invalid list sort", method: http.MethodGet,
			path: "/api/json/workout-routines/4/exercises?sort_by=unknown",
			call: (*WorkoutRoutineExerciseController).GetAll,
		},
		{
			name: "invalid exercise id", method: http.MethodPost,
			path: "/api/json/workout-routines/4/exercises", body: `{"exercise_id":0}`,
			call: (*WorkoutRoutineExerciseController).Create,
		},
		{
			name: "invalid destination", method: http.MethodPatch,
			path: "/api/json/workout-routines/4/exercises/12/order",
			body: `{"destination_position":0}`,
			call: (*WorkoutRoutineExerciseController).Reorder,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			ctx := e.NewContext(request, recorder)
			ctx.SetPathValues(echo.PathValues{
				{Name: "routine_id", Value: "4"},
				{Name: "exercise_id", Value: "12"},
			})
			controller := NewWorkoutRoutineExerciseController(&stubWorkoutRoutineExerciseManager{})
			err := test.call(controller, ctx)
			httpError, ok := err.(*echo.HTTPError)
			if !ok || httpError.Code != http.StatusBadRequest {
				t.Fatalf("error = %v, want HTTP 400", err)
			}
		})
	}
}

func TestWorkoutRoutineExerciseControllerMapsDomainErrors(t *testing.T) {
	tests := []struct {
		err  error
		code int
	}{
		{err: service.ErrWorkoutRoutineNotFound, code: http.StatusNotFound},
		{err: service.ErrExerciseNotFound, code: http.StatusNotFound},
		{err: service.ErrWorkoutRoutineExerciseNotFound, code: http.StatusNotFound},
		{err: service.ErrWorkoutRoutineExerciseConflict, code: http.StatusConflict},
		{err: service.ErrInvalidExercisePosition, code: http.StatusBadRequest},
		{err: service.ErrEmptyExerciseIDs, code: http.StatusBadRequest},
		{err: service.ErrInvalidExerciseID, code: http.StatusBadRequest},
		{err: service.ErrDuplicateExerciseID, code: http.StatusBadRequest},
	}
	for _, test := range tests {
		httpError, ok := workoutRoutineExerciseHTTPError(test.err).(*echo.HTTPError)
		if !ok || httpError.Code != test.code {
			t.Errorf("error = %v, want HTTP %d", httpError, test.code)
		}
	}
}
