package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/labstack/echo/v5"
	"github.com/niviten/heavy/internal/dto"
	"github.com/niviten/heavy/internal/model"
	"github.com/niviten/heavy/internal/service"
)

var validWorkoutRoutineExerciseSortFields = map[string]struct{}{
	"exercise_order":  {},
	"exercise_id":     {},
	"exercise_name":   {},
	"muscle_group_id": {},
}

// WorkoutRoutineExerciseManager is the service behavior required by the controller.
type WorkoutRoutineExerciseManager interface {
	GetAll(
		context.Context,
		int64,
		model.WorkoutRoutineExerciseQuery,
	) (model.WorkoutRoutineExercisePage, error)
	Create(context.Context, int64, int64) (model.WorkoutRoutineExercise, error)
	CreateMany(context.Context, int64, []int64) ([]model.WorkoutRoutineExercise, error)
	Delete(context.Context, int64, int64) error
	Reorder(context.Context, int64, int64, int) (model.WorkoutRoutineExercise, error)
}

// WorkoutRoutineExerciseController serves exercise membership APIs for routines.
type WorkoutRoutineExerciseController struct {
	exercises WorkoutRoutineExerciseManager
}

func NewWorkoutRoutineExerciseController(
	exercises WorkoutRoutineExerciseManager,
) *WorkoutRoutineExerciseController {
	return &WorkoutRoutineExerciseController{exercises: exercises}
}

func (c *WorkoutRoutineExerciseController) GetAll(ctx *echo.Context) error {
	workoutRoutineID, err := positiveID(ctx.Param("routine_id"), "workout routine id")
	if err != nil {
		return err
	}
	query, err := workoutRoutineExerciseQuery(ctx)
	if err != nil {
		return err
	}
	page, err := c.exercises.GetAll(ctx.Request().Context(), workoutRoutineID, query)
	if err != nil {
		return workoutRoutineExerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, workoutRoutineExerciseListResponse(page))
}

func (c *WorkoutRoutineExerciseController) Create(ctx *echo.Context) error {
	workoutRoutineID, err := positiveID(ctx.Param("routine_id"), "workout routine id")
	if err != nil {
		return err
	}
	var request dto.AddWorkoutRoutineExerciseRequest
	if err := ctx.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if request.ExerciseID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "exercise_id must be a positive integer")
	}
	item, err := c.exercises.Create(
		ctx.Request().Context(),
		workoutRoutineID,
		request.ExerciseID,
	)
	if err != nil {
		return workoutRoutineExerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusCreated, workoutRoutineExerciseResponse(item))
}

func (c *WorkoutRoutineExerciseController) CreateMany(ctx *echo.Context) error {
	workoutRoutineID, err := positiveID(ctx.Param("routine_id"), "workout routine id")
	if err != nil {
		return err
	}
	var request dto.BulkAddWorkoutRoutineExercisesRequest
	if err := ctx.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	items, err := c.exercises.CreateMany(
		ctx.Request().Context(),
		workoutRoutineID,
		request.ExerciseIDs,
	)
	if err != nil {
		return workoutRoutineExerciseHTTPError(err)
	}
	responses := make([]dto.WorkoutRoutineExerciseResponse, len(items))
	for index, item := range items {
		responses[index] = workoutRoutineExerciseResponse(item)
	}
	return ctx.JSON(
		http.StatusCreated,
		dto.BulkAddWorkoutRoutineExercisesResponse{Data: responses},
	)
}

func (c *WorkoutRoutineExerciseController) Delete(ctx *echo.Context) error {
	workoutRoutineID, exerciseID, err := workoutRoutineExerciseIDs(ctx)
	if err != nil {
		return err
	}
	if err := c.exercises.Delete(ctx.Request().Context(), workoutRoutineID, exerciseID); err != nil {
		return workoutRoutineExerciseHTTPError(err)
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (c *WorkoutRoutineExerciseController) Reorder(ctx *echo.Context) error {
	workoutRoutineID, exerciseID, err := workoutRoutineExerciseIDs(ctx)
	if err != nil {
		return err
	}
	var request dto.ReorderWorkoutRoutineExerciseRequest
	if err := ctx.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if request.DestinationPosition < 1 {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"destination_position must be a positive integer",
		)
	}
	item, err := c.exercises.Reorder(
		ctx.Request().Context(),
		workoutRoutineID,
		exerciseID,
		request.DestinationPosition,
	)
	if err != nil {
		return workoutRoutineExerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, workoutRoutineExerciseResponse(item))
}

func workoutRoutineExerciseIDs(ctx *echo.Context) (int64, int64, error) {
	workoutRoutineID, err := positiveID(ctx.Param("routine_id"), "workout routine id")
	if err != nil {
		return 0, 0, err
	}
	exerciseID, err := positiveID(ctx.Param("exercise_id"), "exercise id")
	if err != nil {
		return 0, 0, err
	}
	return workoutRoutineID, exerciseID, nil
}

func workoutRoutineExerciseQuery(ctx *echo.Context) (model.WorkoutRoutineExerciseQuery, error) {
	page, err := positiveQueryInteger(ctx.QueryParam("page"), 1, "page")
	if err != nil {
		return model.WorkoutRoutineExerciseQuery{}, err
	}
	pageSize, err := positiveQueryInteger(
		firstNonEmpty(ctx.QueryParam("page_size"), ctx.QueryParam("limit")),
		service.DefaultWorkoutRoutineExercisePageSize,
		"page_size",
	)
	if err != nil {
		return model.WorkoutRoutineExerciseQuery{}, err
	}
	if pageSize > service.MaxWorkoutRoutineExercisePageSize {
		return model.WorkoutRoutineExerciseQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"page_size cannot be greater than 100",
		)
	}

	search := strings.TrimSpace(ctx.QueryParam("search"))
	if utf8.RuneCountInString(search) > 100 {
		return model.WorkoutRoutineExerciseQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"search cannot be greater than 100 characters",
		)
	}
	sortBy := firstNonEmpty(ctx.QueryParam("sort_by"), ctx.QueryParam("sort"))
	if sortBy == "" {
		sortBy = "exercise_order"
	}
	if _, ok := validWorkoutRoutineExerciseSortFields[sortBy]; !ok {
		return model.WorkoutRoutineExerciseQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"sort_by must be exercise_order, exercise_id, exercise_name, or muscle_group_id",
		)
	}
	sortOrder := strings.ToLower(firstNonEmpty(ctx.QueryParam("sort_order"), ctx.QueryParam("order")))
	if sortOrder == "" {
		sortOrder = "asc"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		return model.WorkoutRoutineExerciseQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"sort_order must be asc or desc",
		)
	}
	return model.WorkoutRoutineExerciseQuery{
		Page:      page,
		PageSize:  pageSize,
		Search:    search,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}, nil
}

func workoutRoutineExerciseResponse(
	item model.WorkoutRoutineExercise,
) dto.WorkoutRoutineExerciseResponse {
	return dto.WorkoutRoutineExerciseResponse{
		ID:               item.ID,
		WorkoutRoutineID: item.WorkoutRoutineID,
		ExerciseID:       item.ExerciseID,
		ExerciseName:     item.ExerciseName,
		MuscleGroupID:    item.MuscleGroupID,
		ExerciseOrder:    formatWorkoutRoutineExerciseOrder(item.ExerciseOrder),
	}
}

func workoutRoutineExerciseListResponse(
	page model.WorkoutRoutineExercisePage,
) dto.WorkoutRoutineExerciseListResponse {
	items := make([]dto.WorkoutRoutineExerciseResponse, len(page.Items))
	for index, item := range page.Items {
		items[index] = workoutRoutineExerciseResponse(item)
	}
	return dto.WorkoutRoutineExerciseListResponse{
		Data: items,
		Pagination: dto.PaginationResponse{
			Page:       page.Page,
			PageSize:   page.PageSize,
			TotalItems: page.TotalItems,
			TotalPages: page.TotalPages,
		},
	}
}

func formatWorkoutRoutineExerciseOrder(value int64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	return fmt.Sprintf("%s%d.%03d", sign, value/1000, value%1000)
}

func workoutRoutineExerciseHTTPError(err error) error {
	switch {
	case errors.Is(err, service.ErrWorkoutRoutineNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "workout routine not found")
	case errors.Is(err, service.ErrExerciseNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "exercise not found")
	case errors.Is(err, service.ErrWorkoutRoutineExerciseNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "exercise is not in workout routine")
	case errors.Is(err, service.ErrWorkoutRoutineExerciseConflict):
		return echo.NewHTTPError(http.StatusConflict, "exercise is already in workout routine")
	case errors.Is(err, service.ErrInvalidExercisePosition):
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"destination_position is outside the workout routine",
		)
	case errors.Is(err, service.ErrEmptyExerciseIDs):
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"exercise_ids must contain at least one exercise_id",
		)
	case errors.Is(err, service.ErrInvalidExerciseID):
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"exercise_ids must contain only positive integers",
		)
	case errors.Is(err, service.ErrDuplicateExerciseID):
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"exercise_ids must not contain duplicates",
		)
	default:
		return err
	}
}
