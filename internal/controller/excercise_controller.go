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

var validExcerciseSortFields = map[string]struct{}{
	"excercise_id":    {},
	"excercise_name":  {},
	"muscle_group_id": {},
}

// ExcerciseManager is the service behavior required by the controller.
type ExcerciseManager interface {
	GetAll(context.Context, model.ExcerciseQuery) (model.ExcercisePage, error)
	GetByID(context.Context, int64) (model.Excercise, error)
	GetByMuscleGroup(context.Context, int64, model.ExcerciseQuery) (model.ExcercisePage, error)
	Create(context.Context, string, int64) (model.Excercise, error)
	Update(context.Context, int64, *string, *int64) (model.Excercise, error)
	Delete(context.Context, int64) error
}

// ExcerciseController serves the exercise API. Its name follows the existing
// database spelling.
type ExcerciseController struct {
	excercises ExcerciseManager
}

func NewExcerciseController(excercises ExcerciseManager) *ExcerciseController {
	return &ExcerciseController{excercises: excercises}
}

func (c *ExcerciseController) GetAll(ctx *echo.Context) error {
	query, err := excerciseQuery(ctx)
	if err != nil {
		return err
	}
	page, err := c.excercises.GetAll(ctx.Request().Context(), query)
	if err != nil {
		return excerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, excerciseListResponse(page))
}

func (c *ExcerciseController) GetByID(ctx *echo.Context) error {
	id, err := positiveID(ctx.Param("id"), "excercise id")
	if err != nil {
		return err
	}
	excercise, err := c.excercises.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return excerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, excerciseResponse(excercise))
}

func (c *ExcerciseController) GetByMuscleGroup(ctx *echo.Context) error {
	muscleGroupID, err := positiveID(ctx.Param("muscle_group_id"), "muscle group id")
	if err != nil {
		return err
	}
	query, err := excerciseQuery(ctx)
	if err != nil {
		return err
	}
	page, err := c.excercises.GetByMuscleGroup(ctx.Request().Context(), muscleGroupID, query)
	if err != nil {
		return excerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, excerciseListResponse(page))
}

func (c *ExcerciseController) Create(ctx *echo.Context) error {
	var request dto.CreateExcerciseRequest
	if err := ctx.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if request.MuscleGroupID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "muscle_group_id must be a positive integer")
	}
	excercise, err := c.excercises.Create(
		ctx.Request().Context(),
		request.Name,
		request.MuscleGroupID,
	)
	if err != nil {
		return excerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusCreated, excerciseResponse(excercise))
}

func (c *ExcerciseController) Update(ctx *echo.Context) error {
	id, err := positiveID(ctx.Param("id"), "excercise id")
	if err != nil {
		return err
	}
	var request dto.UpdateExcerciseRequest
	if err := ctx.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	excercise, err := c.excercises.Update(
		ctx.Request().Context(),
		id,
		request.Name,
		request.MuscleGroupID,
	)
	if err != nil {
		return excerciseHTTPError(err)
	}
	return ctx.JSON(http.StatusOK, excerciseResponse(excercise))
}

func (c *ExcerciseController) Delete(ctx *echo.Context) error {
	id, err := positiveID(ctx.Param("id"), "excercise id")
	if err != nil {
		return err
	}
	if err := c.excercises.Delete(ctx.Request().Context(), id); err != nil {
		return excerciseHTTPError(err)
	}
	return ctx.NoContent(http.StatusNoContent)
}

func excerciseQuery(ctx *echo.Context) (model.ExcerciseQuery, error) {
	page, err := positiveQueryInteger(ctx.QueryParam("page"), 1, "page")
	if err != nil {
		return model.ExcerciseQuery{}, err
	}
	pageSize, err := positiveQueryInteger(
		firstNonEmpty(ctx.QueryParam("page_size"), ctx.QueryParam("limit")),
		service.DefaultExcercisePageSize,
		"page_size",
	)
	if err != nil {
		return model.ExcerciseQuery{}, err
	}
	if pageSize > service.MaxExcercisePageSize {
		return model.ExcerciseQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"page_size cannot be greater than 100",
		)
	}

	search := strings.TrimSpace(ctx.QueryParam("search"))
	if utf8.RuneCountInString(search) > 100 {
		return model.ExcerciseQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"search cannot be greater than 100 characters",
		)
	}
	sortBy := firstNonEmpty(ctx.QueryParam("sort_by"), ctx.QueryParam("sort"))
	if sortBy == "" {
		sortBy = "excercise_id"
	}
	if _, ok := validExcerciseSortFields[sortBy]; !ok {
		return model.ExcerciseQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"sort_by must be excercise_id, excercise_name, or muscle_group_id",
		)
	}
	sortOrder := strings.ToLower(firstNonEmpty(ctx.QueryParam("sort_order"), ctx.QueryParam("order")))
	if sortOrder == "" {
		sortOrder = "asc"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		return model.ExcerciseQuery{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"sort_order must be asc or desc",
		)
	}

	return model.ExcerciseQuery{
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

func excerciseResponse(excercise model.Excercise) dto.ExcerciseResponse {
	return dto.ExcerciseResponse{
		ID:            excercise.ID,
		Name:          excercise.Name,
		MuscleGroupID: excercise.MuscleGroupID,
	}
}

func excerciseListResponse(page model.ExcercisePage) dto.ExcerciseListResponse {
	items := make([]dto.ExcerciseResponse, len(page.Items))
	for index, excercise := range page.Items {
		items[index] = excerciseResponse(excercise)
	}
	return dto.ExcerciseListResponse{
		Data: items,
		Pagination: dto.PaginationResponse{
			Page:       page.Page,
			PageSize:   page.PageSize,
			TotalItems: page.TotalItems,
			TotalPages: page.TotalPages,
		},
	}
}

func excerciseHTTPError(err error) error {
	switch {
	case errors.Is(err, service.ErrExcerciseNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "excercise not found")
	case errors.Is(err, service.ErrMuscleGroupNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "muscle group not found")
	case errors.Is(err, service.ErrExcerciseNameConflict):
		return echo.NewHTTPError(http.StatusConflict, "excercise name already exists")
	case errors.Is(err, service.ErrInvalidExcerciseName):
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"excercise_name must contain between 1 and 100 characters",
		)
	case errors.Is(err, service.ErrInvalidMuscleGroupID):
		return echo.NewHTTPError(http.StatusBadRequest, "muscle_group_id must be a positive integer")
	case errors.Is(err, service.ErrEmptyExcerciseUpdate):
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"provide excercise_name, muscle_group_id, or both",
		)
	default:
		return err
	}
}
