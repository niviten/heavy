package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/niviten/heavy/internal/model"
	"github.com/niviten/heavy/internal/service"
)

type stubMuscleGroupReader struct {
	muscleGroups []model.MuscleGroup
	muscleGroup  model.MuscleGroup
	err          error
}

func (s stubMuscleGroupReader) GetAll(context.Context) ([]model.MuscleGroup, error) {
	return s.muscleGroups, s.err
}

func (s stubMuscleGroupReader) GetByID(context.Context, int64) (model.MuscleGroup, error) {
	return s.muscleGroup, s.err
}

func (s stubMuscleGroupReader) GetByName(context.Context, string) (model.MuscleGroup, error) {
	return s.muscleGroup, s.err
}

func TestMuscleGroupControllerGetAll(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/json/muscle-groups", nil)
	ctx := e.NewContext(request, recorder)
	controller := NewMuscleGroupController(stubMuscleGroupReader{muscleGroups: []model.MuscleGroup{{
		ID:          1,
		Name:        "chest",
		DisplayName: "Chest",
	}}})

	if err := controller.GetAll(ctx); err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Body.String(); got != "[{\"muscle_group_id\":1,\"muscle_group_name\":\"chest\",\"muscle_group_display_name\":\"Chest\"}]\n" {
		t.Errorf("body = %q", got)
	}
}

func TestMuscleGroupControllerGetByIDRejectsInvalidID(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/json/muscle-groups/nope", nil)
	ctx := e.NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "nope"}})
	controller := NewMuscleGroupController(stubMuscleGroupReader{})

	err := controller.GetByID(ctx)
	httpError, ok := err.(*echo.HTTPError)
	if !ok || httpError.Code != http.StatusBadRequest {
		t.Fatalf("GetByID() error = %v, want HTTP 400", err)
	}
}

func TestMuscleGroupControllerReturnsNotFound(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/json/muscle-groups/99", nil)
	ctx := e.NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})
	controller := NewMuscleGroupController(stubMuscleGroupReader{err: service.ErrMuscleGroupNotFound})

	err := controller.GetByID(ctx)
	httpError, ok := err.(*echo.HTTPError)
	if !ok || httpError.Code != http.StatusNotFound {
		t.Fatalf("GetByID() error = %v, want HTTP 404", err)
	}
}
