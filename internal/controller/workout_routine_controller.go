package controller

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/labstack/echo/v5"
	"github.com/niviten/heavy/internal/dto"
	"github.com/niviten/heavy/internal/model"
	"github.com/niviten/heavy/internal/service"
)

var validWorkoutRoutineSortFields = map[string]struct{}{
	"workout_routine_id":          {},
	"workout_routine_name":        {},
	"workout_routine_description": {},
}

// WorkoutRoutineManager is the service behavior required by the controller.
type WorkoutRoutineManager interface {
	GetAll(context.Context, model.WorkoutRoutineQuery) (model.WorkoutRoutinePage, error)
	GetByID(context.Context, int64) (model.WorkoutRoutine, error)
	Create(context.Context, string, *string) (model.WorkoutRoutine, error)
	Update(context.Context, int64, *string, *string) (model.WorkoutRoutine, error)
	Delete(context.Context, int64) error
}

// WorkoutRoutineController serves the workout routine API.
type WorkoutRoutineController struct {
	routines WorkoutRoutineManager
}

func NewWorkoutRoutineController(routines WorkoutRoutineManager) *WorkoutRoutineController {
	return &WorkoutRoutineController{routines: routines}
}

func (c *WorkoutRoutineController) GetAll(ctx *echo.Context) error {
	query, err := workoutRoutineQuery(ctx)
	if err != nil {
		return err
	}
	page, err := c.routines.GetAll(ctx.Request().Context(), query)
	if err != nil {
		return workoutRoutineHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, workoutRoutineListResponse(page))
}

func (c *WorkoutRoutineController) GetByID(ctx *echo.Context) error {
	id, err := positiveID(ctx.Param("id"), "workout routine id")
	if err != nil {
		return err
	}
	routine, err := c.routines.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return workoutRoutineHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, workoutRoutineResponse(routine))
}

func (c *WorkoutRoutineController) Create(ctx *echo.Context) error {
	var request dto.CreateWorkoutRoutineRequest
	if err := ctx.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	routine, err := c.routines.Create(
		ctx.Request().Context(),
		request.Name,
		request.Description,
	)
	if err != nil {
		return workoutRoutineHTTPError(err)
	}
	return ctx.JSON(http.StatusCreated, workoutRoutineResponse(routine))
}

func (c *WorkoutRoutineController) Update(ctx *echo.Context) error {
	id, err := positiveID(ctx.Param("id"), "workout routine id")
	if err != nil {
		return err
	}
	var request dto.UpdateWorkoutRoutineRequest
	if err := ctx.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	routine, err := c.routines.Update(
		ctx.Request().Context(),
		id,
		request.Name,
		request.Description,
	)
	if err != nil {
		return workoutRoutineHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, workoutRoutineResponse(routine))
}

func (c *WorkoutRoutineController) Delete(ctx *echo.Context) error {
	id, err := positiveID(ctx.Param("id"), "workout routine id")
	if err != nil {
		return err
	}
	if err := c.routines.Delete(ctx.Request().Context(), id); err != nil {
		return workoutRoutineHTTPError(err)
	}
	return ctx.NoContent(http.StatusNoContent)
}

func workoutRoutineQuery(ctx *echo.Context) (model.WorkoutRoutineQuery, error) {
	page, err := positiveQueryInteger(ctx.QueryParam("page"), 1, "page")
	if err != nil {
		return model.WorkoutRoutineQuery{}, err
	}
	pageSize, err := positiveQueryInteger(
		firstNonEmpty(ctx.QueryParam("page_size"), ctx.QueryParam("limit")),
		service.DefaultWorkoutRoutinePageSize,
		"page_size",
	)
	if err != nil {
		return model.WorkoutRoutineQuery{}, err
	}
	if pageSize > service.MaxWorkoutRoutinePageSize {
		return model.WorkoutRoutineQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"page_size cannot be greater than 100",
		)
	}

	search := strings.TrimSpace(ctx.QueryParam("search"))
	if utf8.RuneCountInString(search) > 255 {
		return model.WorkoutRoutineQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"search cannot be greater than 255 characters",
		)
	}
	sortBy := firstNonEmpty(ctx.QueryParam("sort_by"), ctx.QueryParam("sort"))
	if sortBy == "" {
		sortBy = "workout_routine_id"
	}
	if _, ok := validWorkoutRoutineSortFields[sortBy]; !ok {
		return model.WorkoutRoutineQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"sort_by must be workout_routine_id, workout_routine_name, or workout_routine_description",
		)
	}
	sortOrder := strings.ToLower(firstNonEmpty(ctx.QueryParam("sort_order"), ctx.QueryParam("order")))
	if sortOrder == "" {
		sortOrder = "asc"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		return model.WorkoutRoutineQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"sort_order must be asc or desc",
		)
	}

	return model.WorkoutRoutineQuery{
		Page:      page,
		PageSize:  pageSize,
		Search:    search,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}, nil
}

func workoutRoutineResponse(routine model.WorkoutRoutine) dto.WorkoutRoutineResponse {
	return dto.WorkoutRoutineResponse{
		ID:          routine.ID,
		Name:        routine.Name,
		Description: routine.Description,
	}
}

func workoutRoutineListResponse(page model.WorkoutRoutinePage) dto.WorkoutRoutineListResponse {
	items := make([]dto.WorkoutRoutineResponse, len(page.Items))
	for index, routine := range page.Items {
		items[index] = workoutRoutineResponse(routine)
	}
	return dto.WorkoutRoutineListResponse{
		Data: items,
		Pagination: dto.PaginationResponse{
			Page:       page.Page,
			PageSize:   page.PageSize,
			TotalItems: page.TotalItems,
			TotalPages: page.TotalPages,
		},
	}
}

func workoutRoutineHTTPError(err error) error {
	switch {
	case errors.Is(err, service.ErrWorkoutRoutineNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "workout routine not found")
	case errors.Is(err, service.ErrWorkoutRoutineNameConflict):
		return echo.NewHTTPError(http.StatusConflict, "workout routine name already exists")
	case errors.Is(err, service.ErrInvalidWorkoutRoutineName):
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"workout_routine_name must contain between 1 and 100 characters",
		)
	case errors.Is(err, service.ErrInvalidWorkoutRoutineDescription):
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"workout_routine_description cannot exceed 255 characters",
		)
	case errors.Is(err, service.ErrEmptyWorkoutRoutineUpdate):
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"provide workout_routine_name, workout_routine_description, or both",
		)
	default:
		return err
	}
}
