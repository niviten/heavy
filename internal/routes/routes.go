package routes

import (
	"github.com/labstack/echo/v5"
	"github.com/niviten/heavy/internal/controller"
)

// Register attaches all application routes to Echo.
func Register(
	e *echo.Echo,
	health *controller.HealthController,
	muscleGroups *controller.MuscleGroupController,
) {
	api := e.Group("/api/json")

	api.GET("", health.Root)
	api.GET("/health", health.Health)
	api.GET("/muscle-groups", muscleGroups.GetAll)
	api.GET("/muscle-groups/name/:name", muscleGroups.GetByName)
	api.GET("/muscle-groups/:id", muscleGroups.GetByID)
}
