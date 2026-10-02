package routes

import (
	"github.com/labstack/echo/v5"
	"github.com/niviten/heavy/internal/controller"
)

// Register attaches all application routes to Echo.
func Register(e *echo.Echo, health *controller.HealthController) {
	e.GET("/", health.Root)
	e.GET("/health", health.Health)
}
