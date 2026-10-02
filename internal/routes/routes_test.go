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

func TestRegister(t *testing.T) {
	e := echo.New()
	Register(
		e,
		controller.NewHealthController(healthyChecker{}),
		controller.NewMuscleGroupController(muscleGroupReader{}),
	)

	for _, path := range []string{
		"/api/json",
		"/api/json/health",
		"/api/json/muscle-groups",
		"/api/json/muscle-groups/1",
		"/api/json/muscle-groups/name/Chest",
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
