package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/niviten/heavy/internal/controller"
	"github.com/niviten/heavy/internal/dto"
)

type healthyChecker struct{}

func (healthyChecker) Check(context.Context) (dto.HealthResponse, error) {
	return dto.HealthResponse{Status: "ok", Database: "up"}, nil
}

func TestRegister(t *testing.T) {
	e := echo.New()
	Register(e, controller.NewHealthController(healthyChecker{}))

	for _, path := range []string{"/", "/health"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		e.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}
	}
}
