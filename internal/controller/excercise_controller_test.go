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

type stubExcerciseManager struct {
	page       model.ExcercisePage
	excercise  model.Excercise
	query      model.ExcerciseQuery
	id         int64
	muscleID   int64
	name       string
	err        error
	deleteCall bool
}

func (s *stubExcerciseManager) GetAll(
	_ context.Context,
	query model.ExcerciseQuery,
) (model.ExcercisePage, error) {
	s.query = query
	return s.page, s.err
}

func (s *stubExcerciseManager) GetByID(_ context.Context, id int64) (model.Excercise, error) {
	s.id = id
	return s.excercise, s.err
}

func (s *stubExcerciseManager) GetByMuscleGroup(
	_ context.Context,
	muscleGroupID int64,
	query model.ExcerciseQuery,
) (model.ExcercisePage, error) {
	s.muscleID = muscleGroupID
	s.query = query
	return s.page, s.err
}

func (s *stubExcerciseManager) Create(
	_ context.Context,
	name string,
	muscleGroupID int64,
) (model.Excercise, error) {
	s.name = name
	s.muscleID = muscleGroupID
	return s.excercise, s.err
}

func (s *stubExcerciseManager) Update(
	_ context.Context,
	id int64,
	name *string,
	muscleGroupID *int64,
) (model.Excercise, error) {
	s.id = id
	if name != nil {
		s.name = *name
	}
	if muscleGroupID != nil {
		s.muscleID = *muscleGroupID
	}
	return s.excercise, s.err
}

func (s *stubExcerciseManager) Delete(_ context.Context, id int64) error {
	s.id = id
	s.deleteCall = true
	return s.err
}

func TestExcerciseControllerGetAll(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/json/excercises?page=2&page_size=5&search=press&sort_by=excercise_name&sort_order=desc",
		nil,
	)
	ctx := e.NewContext(request, recorder)
	manager := &stubExcerciseManager{page: model.ExcercisePage{
		Items:      []model.Excercise{{ID: 3, Name: "Bench Press", MuscleGroupID: 1}},
		Page:       2,
		PageSize:   5,
		TotalItems: 8,
		TotalPages: 2,
	}}
	controller := NewExcerciseController(manager)

	if err := controller.GetAll(ctx); err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	wantBody := "{\"data\":[{\"excercise_id\":3,\"excercise_name\":\"Bench Press\",\"muscle_group_id\":1}]," +
		"\"pagination\":{\"page\":2,\"page_size\":5,\"total_items\":8,\"total_pages\":2}}\n"
	if got := recorder.Body.String(); got != wantBody {
		t.Errorf("body = %q, want %q", got, wantBody)
	}
	if manager.query.Page != 2 || manager.query.PageSize != 5 || manager.query.Search != "press" ||
		manager.query.SortBy != "excercise_name" || manager.query.SortOrder != "desc" {
		t.Errorf("query = %+v", manager.query)
	}
}

func TestExcerciseControllerRejectsInvalidCollectionQuery(t *testing.T) {
	tests := []string{
		"?page=0",
		"?page_size=101",
		"?sort_by=is_deleted",
		"?sort_order=sideways",
	}
	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			e := echo.New()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/json/excercises"+query, nil)
			ctx := e.NewContext(request, recorder)
			controller := NewExcerciseController(&stubExcerciseManager{})

			err := controller.GetAll(ctx)
			httpError, ok := err.(*echo.HTTPError)
			if !ok || httpError.Code != http.StatusBadRequest {
				t.Fatalf("GetAll() error = %v, want HTTP 400", err)
			}
		})
	}
}

func TestExcerciseControllerCreate(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/json/excercises",
		strings.NewReader(`{"excercise_name":"Bench Press","muscle_group_id":1}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	ctx := e.NewContext(request, recorder)
	manager := &stubExcerciseManager{excercise: model.Excercise{
		ID: 4, Name: "Bench Press", MuscleGroupID: 1,
	}}
	controller := NewExcerciseController(manager)

	if err := controller.Create(ctx); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if recorder.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if manager.name != "Bench Press" || manager.muscleID != 1 {
		t.Errorf("name = %q, muscle group = %d", manager.name, manager.muscleID)
	}
}

func TestExcerciseControllerDelete(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/api/json/excercises/4", nil)
	ctx := e.NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "4"}})
	manager := &stubExcerciseManager{}
	controller := NewExcerciseController(manager)

	if err := controller.Delete(ctx); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if recorder.Code != http.StatusNoContent || !manager.deleteCall || manager.id != 4 {
		t.Errorf("status = %d, called = %v, id = %d", recorder.Code, manager.deleteCall, manager.id)
	}
}

func TestExcerciseControllerUpdatesMuscleGroup(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/json/excercises/4",
		strings.NewReader(`{"muscle_group_id":2}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	ctx := e.NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "4"}})
	manager := &stubExcerciseManager{excercise: model.Excercise{
		ID: 4, Name: "Bench Press", MuscleGroupID: 2,
	}}
	controller := NewExcerciseController(manager)

	if err := controller.Update(ctx); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if recorder.Code != http.StatusOK || manager.id != 4 || manager.muscleID != 2 {
		t.Errorf("status = %d, id = %d, muscle group = %d", recorder.Code, manager.id, manager.muscleID)
	}
}

func TestExcerciseControllerMapsDomainErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code int
	}{
		{name: "not found", err: service.ErrExcerciseNotFound, code: http.StatusNotFound},
		{name: "muscle group not found", err: service.ErrMuscleGroupNotFound, code: http.StatusNotFound},
		{name: "duplicate", err: service.ErrExcerciseNameConflict, code: http.StatusConflict},
		{name: "invalid name", err: service.ErrInvalidExcerciseName, code: http.StatusBadRequest},
		{name: "invalid muscle group id", err: service.ErrInvalidMuscleGroupID, code: http.StatusBadRequest},
		{name: "empty update", err: service.ErrEmptyExcerciseUpdate, code: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpError, ok := excerciseHTTPError(test.err).(*echo.HTTPError)
			if !ok || httpError.Code != test.code {
				t.Fatalf("error = %v, want HTTP %d", httpError, test.code)
			}
		})
	}
}
