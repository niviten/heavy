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

type stubWorkoutRoutineManager struct {
	page        model.WorkoutRoutinePage
	routine     model.WorkoutRoutine
	query       model.WorkoutRoutineQuery
	id          int64
	name        string
	description *string
	err         error
	deleteCall  bool
}

func (s *stubWorkoutRoutineManager) GetAll(
	_ context.Context,
	query model.WorkoutRoutineQuery,
) (model.WorkoutRoutinePage, error) {
	s.query = query
	return s.page, s.err
}

func (s *stubWorkoutRoutineManager) GetByID(
	_ context.Context,
	id int64,
) (model.WorkoutRoutine, error) {
	s.id = id
	return s.routine, s.err
}

func (s *stubWorkoutRoutineManager) Create(
	_ context.Context,
	name string,
	description *string,
) (model.WorkoutRoutine, error) {
	s.name = name
	s.description = description
	return s.routine, s.err
}

func (s *stubWorkoutRoutineManager) Update(
	_ context.Context,
	id int64,
	name *string,
	description *string,
) (model.WorkoutRoutine, error) {
	s.id = id
	if name != nil {
		s.name = *name
	}
	s.description = description
	return s.routine, s.err
}

func (s *stubWorkoutRoutineManager) Delete(_ context.Context, id int64) error {
	s.id = id
	s.deleteCall = true
	return s.err
}

func TestWorkoutRoutineControllerGetAll(t *testing.T) {
	description := "Chest and shoulders"
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/json/workout-routines?page=2&page_size=5&search=push&sort_by=workout_routine_name&sort_order=desc",
		nil,
	)
	ctx := e.NewContext(request, recorder)
	manager := &stubWorkoutRoutineManager{page: model.WorkoutRoutinePage{
		Items:      []model.WorkoutRoutine{{ID: 3, Name: "Push Day", Description: &description}},
		Page:       2,
		PageSize:   5,
		TotalItems: 8,
		TotalPages: 2,
	}}
	controller := NewWorkoutRoutineController(manager)

	if err := controller.GetAll(ctx); err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	wantBody := "{\"data\":[{\"workout_routine_id\":3,\"workout_routine_name\":\"Push Day\"," +
		"\"workout_routine_description\":\"Chest and shoulders\"}]," +
		"\"pagination\":{\"page\":2,\"page_size\":5,\"total_items\":8,\"total_pages\":2}}\n"
	if got := recorder.Body.String(); got != wantBody {
		t.Errorf("body = %q, want %q", got, wantBody)
	}
	if manager.query.Page != 2 || manager.query.PageSize != 5 || manager.query.Search != "push" ||
		manager.query.SortBy != "workout_routine_name" || manager.query.SortOrder != "desc" {
		t.Errorf("query = %+v", manager.query)
	}
}

func TestWorkoutRoutineControllerRejectsInvalidCollectionQuery(t *testing.T) {
	tests := []string{
		"?page=0",
		"?page_size=101",
		"?sort_by=unknown",
		"?sort_order=sideways",
	}
	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			e := echo.New()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/json/workout-routines"+query, nil)
			ctx := e.NewContext(request, recorder)
			controller := NewWorkoutRoutineController(&stubWorkoutRoutineManager{})

			err := controller.GetAll(ctx)
			httpError, ok := err.(*echo.HTTPError)
			if !ok || httpError.Code != http.StatusBadRequest {
				t.Fatalf("GetAll() error = %v, want HTTP 400", err)
			}
		})
	}
}

func TestWorkoutRoutineControllerCreate(t *testing.T) {
	description := "Chest and shoulders"
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/json/workout-routines",
		strings.NewReader(`{"workout_routine_name":"Push Day","workout_routine_description":"Chest and shoulders"}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	ctx := e.NewContext(request, recorder)
	manager := &stubWorkoutRoutineManager{routine: model.WorkoutRoutine{
		ID: 1, Name: "Push Day", Description: &description,
	}}
	controller := NewWorkoutRoutineController(manager)

	if err := controller.Create(ctx); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if recorder.Code != http.StatusCreated || manager.name != "Push Day" {
		t.Errorf("status = %d, name = %q", recorder.Code, manager.name)
	}
	if manager.description == nil || *manager.description != description {
		t.Errorf("description = %v", manager.description)
	}
}

func TestWorkoutRoutineControllerUpdateNameAndDescription(t *testing.T) {
	description := "Chest, shoulders, and triceps"
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/json/workout-routines/4",
		strings.NewReader(`{"workout_routine_name":"Upper Push","workout_routine_description":"Chest, shoulders, and triceps"}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	ctx := e.NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "4"}})
	manager := &stubWorkoutRoutineManager{routine: model.WorkoutRoutine{
		ID: 4, Name: "Upper Push", Description: &description,
	}}
	controller := NewWorkoutRoutineController(manager)

	if err := controller.Update(ctx); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if recorder.Code != http.StatusOK || manager.id != 4 || manager.name != "Upper Push" {
		t.Errorf("status = %d, id = %d, name = %q", recorder.Code, manager.id, manager.name)
	}
	if manager.description == nil || *manager.description != description {
		t.Errorf("description = %v", manager.description)
	}
}

func TestWorkoutRoutineControllerDelete(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/api/json/workout-routines/4", nil)
	ctx := e.NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "4"}})
	manager := &stubWorkoutRoutineManager{}
	controller := NewWorkoutRoutineController(manager)

	if err := controller.Delete(ctx); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if recorder.Code != http.StatusNoContent || !manager.deleteCall || manager.id != 4 {
		t.Errorf("status = %d, called = %v, id = %d", recorder.Code, manager.deleteCall, manager.id)
	}
}

func TestWorkoutRoutineControllerMapsDomainErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code int
	}{
		{name: "not found", err: service.ErrWorkoutRoutineNotFound, code: http.StatusNotFound},
		{name: "duplicate", err: service.ErrWorkoutRoutineNameConflict, code: http.StatusConflict},
		{name: "invalid name", err: service.ErrInvalidWorkoutRoutineName, code: http.StatusBadRequest},
		{name: "invalid description", err: service.ErrInvalidWorkoutRoutineDescription, code: http.StatusBadRequest},
		{name: "empty update", err: service.ErrEmptyWorkoutRoutineUpdate, code: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpError, ok := workoutRoutineHTTPError(test.err).(*echo.HTTPError)
			if !ok || httpError.Code != test.code {
				t.Fatalf("error = %v, want HTTP %d", httpError, test.code)
			}
		})
	}
}
