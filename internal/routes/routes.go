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
	exercises *controller.ExerciseController,
	workoutRoutines *controller.WorkoutRoutineController,
) {
	api := e.Group("/api/json")

	api.GET("", health.Root)
	api.GET("/health", health.Health)
	api.GET("/muscle-groups", muscleGroups.GetAll)
	api.GET("/muscle-groups/name/:name", muscleGroups.GetByName)
	api.GET("/muscle-groups/:id", muscleGroups.GetByID)
	api.GET("/muscle-groups/:muscle_group_id/exercises", exercises.GetByMuscleGroup)

	api.GET("/exercises", exercises.GetAll)
	api.POST("/exercises", exercises.Create)
	api.GET("/exercises/:id", exercises.GetByID)
	api.PATCH("/exercises/:id", exercises.Update)
	api.DELETE("/exercises/:id", exercises.Delete)

	api.GET("/workout-routines", workoutRoutines.GetAll)
	api.POST("/workout-routines", workoutRoutines.Create)
	api.GET("/workout-routines/:id", workoutRoutines.GetByID)
	api.PATCH("/workout-routines/:id", workoutRoutines.Update)
	api.DELETE("/workout-routines/:id", workoutRoutines.Delete)
}
