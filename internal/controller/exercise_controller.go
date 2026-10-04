package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/labstack/echo/v5"
	"github.com/niviten/heavy/internal/dto"
	"github.com/niviten/heavy/internal/model"
	"github.com/niviten/heavy/internal/service"
)

var validExerciseSortFields = map[string]struct{}{
	"exercise_id":     {},
	"exercise_name":   {},
	"muscle_group_id": {},
}

// ExerciseManager is the service behavior required by the controller.
type ExerciseManager interface {
	GetAll(context.Context, model.ExerciseQuery) (model.ExercisePage, error)
	GetByID(context.Context, int64) (model.Exercise, error)
	GetByMuscleGroup(context.Context, int64, model.ExerciseQuery) (model.ExercisePage, error)
	Create(context.Context, string, int64) (model.Exercise, error)
	Update(context.Context, int64, *string, *int64) (model.Exercise, error)
	Delete(context.Context, int64) error
}

// ExerciseController serves the exercise API. Its name follows the existing
// database spelling.
type ExerciseController struct {
	exercises ExerciseManager
}

func NewExerciseController(exercises ExerciseManager) *ExerciseController {
	return &ExerciseController{exercises: exercises}
}

func (c *ExerciseController) GetAll(ctx *echo.Context) error {
	query, err := exerciseQuery(ctx)
	if err != nil {
		return err
	}
	page, err := c.exercises.GetAll(ctx.Request().Context(), query)
	if err != nil {
		return exerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, exerciseListResponse(page))
}

func (c *ExerciseController) GetByID(ctx *echo.Context) error {
	id, err := positiveID(ctx.Param("id"), "exercise id")
	if err != nil {
		return err
	}
	exercise, err := c.exercises.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return exerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, exerciseResponse(exercise))
}

func (c *ExerciseController) GetByMuscleGroup(ctx *echo.Context) error {
	muscleGroupID, err := positiveID(ctx.Param("muscle_group_id"), "muscle group id")
	if err != nil {
		return err
	}
	query, err := exerciseQuery(ctx)
	if err != nil {
		return err
	}
	page, err := c.exercises.GetByMuscleGroup(ctx.Request().Context(), muscleGroupID, query)
	if err != nil {
		return exerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, exerciseListResponse(page))
}

func (c *ExerciseController) Create(ctx *echo.Context) error {
	var request dto.CreateExerciseRequest
	if err := ctx.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if request.MuscleGroupID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "muscle_group_id must be a positive integer")
	}
	exercise, err := c.exercises.Create(
		ctx.Request().Context(),
		request.Name,
		request.MuscleGroupID,
	)
	if err != nil {
		return exerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusCreated, exerciseResponse(exercise))
}

func (c *ExerciseController) Update(ctx *echo.Context) error {
	id, err := positiveID(ctx.Param("id"), "exercise id")
	if err != nil {
		return err
	}
	var request dto.UpdateExerciseRequest
	if err := ctx.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	exercise, err := c.exercises.Update(
		ctx.Request().Context(),
		id,
		request.Name,
		request.MuscleGroupID,
	)
	if err != nil {
		return exerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, exerciseResponse(exercise))
}

func (c *ExerciseController) Delete(ctx *echo.Context) error {
	id, err := positiveID(ctx.Param("id"), "exercise id")
	if err != nil {
		return err
	}
	if err := c.exercises.Delete(ctx.Request().Context(), id); err != nil {
		return exerciseHTTPError(err)
	}
	return ctx.NoContent(http.StatusNoContent)
}

func exerciseQuery(ctx *echo.Context) (model.ExerciseQuery, error) {
	page, err := positiveQueryInteger(ctx.QueryParam("page"), 1, "page")
	if err != nil {
		return model.ExerciseQuery{}, err
	}
	pageSize, err := positiveQueryInteger(
		firstNonEmpty(ctx.QueryParam("page_size"), ctx.QueryParam("limit")),
		service.DefaultExercisePageSize,
		"page_size",
	)
	if err != nil {
		return model.ExerciseQuery{}, err
	}
	if pageSize > service.MaxExercisePageSize {
		return model.ExerciseQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"page_size cannot be greater than 100",
		)
	}

	search := strings.TrimSpace(ctx.QueryParam("search"))
	if utf8.RuneCountInString(search) > 100 {
		return model.ExerciseQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"search cannot be greater than 100 characters",
		)
	}
	sortBy := firstNonEmpty(ctx.QueryParam("sort_by"), ctx.QueryParam("sort"))
	if sortBy == "" {
		sortBy = "exercise_id"
	}
	if _, ok := validExerciseSortFields[sortBy]; !ok {
		return model.ExerciseQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"sort_by must be exercise_id, exercise_name, or muscle_group_id",
		)
	}
	sortOrder := strings.ToLower(firstNonEmpty(ctx.QueryParam("sort_order"), ctx.QueryParam("order")))
	if sortOrder == "" {
		sortOrder = "asc"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		return model.ExerciseQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"sort_order must be asc or desc",
		)
	}

	return model.ExerciseQuery{
		Page:      page,
		PageSize:  pageSize,
		Search:    search,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}, nil
}

func positiveQueryInteger(raw string, defaultValue int, name string) (int, error) {
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, name+" must be a positive integer")
	}
	return value, nil
}

func positiveID(raw string, name string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, name+" must be a positive integer")
	}
	return id, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func exerciseResponse(exercise model.Exercise) dto.ExerciseResponse {
	return dto.ExerciseResponse{
		ID:            exercise.ID,
		Name:          exercise.Name,
		MuscleGroupID: exercise.MuscleGroupID,
	}
}

func exerciseListResponse(page model.ExercisePage) dto.ExerciseListResponse {
	items := make([]dto.ExerciseResponse, len(page.Items))
	for index, exercise := range page.Items {
		items[index] = exerciseResponse(exercise)
	}
	return dto.ExerciseListResponse{
		Data: items,
		Pagination: dto.PaginationResponse{
			Page:       page.Page,
			PageSize:   page.PageSize,
			TotalItems: page.TotalItems,
			TotalPages: page.TotalPages,
		},
	}
}

func exerciseHTTPError(err error) error {
	switch {
	case errors.Is(err, service.ErrExerciseNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "exercise not found")
	case errors.Is(err, service.ErrMuscleGroupNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "muscle group not found")
	case errors.Is(err, service.ErrExerciseNameConflict):
		return echo.NewHTTPError(http.StatusConflict, "exercise name already exists")
	case errors.Is(err, service.ErrInvalidExerciseName):
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"exercise_name must contain between 1 and 100 characters",
		)
	case errors.Is(err, service.ErrInvalidMuscleGroupID):
		return echo.NewHTTPError(http.StatusBadRequest, "muscle_group_id must be a positive integer")
	case errors.Is(err, service.ErrEmptyExerciseUpdate):
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"provide exercise_name, muscle_group_id, or both",
		)
	default:
		return err
	}
}
