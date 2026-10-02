package controller

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/niviten/heavy/internal/dto"
)

// HealthChecker is the service behavior required by the controller.
type HealthChecker interface {
	Check(context.Context) (dto.HealthResponse, error)
}

// HealthController handles the bootstrap HTTP endpoints.
type HealthController struct {
	health HealthChecker
}

func NewHealthController(health HealthChecker) *HealthController {
	return &HealthController{health: health}
}

func (h *HealthController) Root(c *echo.Context) error {
	return c.JSON(http.StatusOK, dto.MessageResponse{Message: "heavy API"})
}

func (h *HealthController) Health(c *echo.Context) error {
	result, err := h.health.Check(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, result)
	}

	return c.JSON(http.StatusOK, result)
}
