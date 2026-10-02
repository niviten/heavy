package controller

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/niviten/heavy/internal/dto"
)

type stubHealthChecker struct {
	result dto.HealthResponse
	err    error
}

func (s stubHealthChecker) Check(context.Context) (dto.HealthResponse, error) {
	return s.result, s.err
}

func TestRoot(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := e.NewContext(request, recorder)
	controller := NewHealthController(stubHealthChecker{})

	if err := controller.Root(ctx); err != nil {
		t.Fatalf("Root() error = %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `"message":"heavy API"`) {
		t.Errorf("body = %q, want application message", recorder.Body.String())
	}
}

func TestHealth(t *testing.T) {
	tests := []struct {
		name       string
		checker    stubHealthChecker
		wantStatus int
	}{
		{
			name:       "healthy",
			checker:    stubHealthChecker{result: dto.HealthResponse{Status: "ok", Database: "up"}},
			wantStatus: http.StatusOK,
		},
		{
			name: "unhealthy",
			checker: stubHealthChecker{
				result: dto.HealthResponse{Status: "unhealthy", Database: "down"},
				err:    errors.New("database unavailable"),
			},
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/health", nil)
			ctx := e.NewContext(request, recorder)

			if err := NewHealthController(tt.checker).Health(ctx); err != nil {
				t.Fatalf("Health() error = %v", err)
			}
			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}
