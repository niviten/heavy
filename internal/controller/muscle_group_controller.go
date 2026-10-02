package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/niviten/heavy/internal/dto"
	"github.com/niviten/heavy/internal/model"
	"github.com/niviten/heavy/internal/service"
)

// MuscleGroupReader is the service behavior required by the controller.
type MuscleGroupReader interface {
	GetAll(context.Context) ([]model.MuscleGroup, error)
	GetByID(context.Context, int64) (model.MuscleGroup, error)
	GetByName(context.Context, string) (model.MuscleGroup, error)
}

// MuscleGroupController serves the read-only muscle group API.
type MuscleGroupController struct {
	muscleGroups MuscleGroupReader
}

func NewMuscleGroupController(muscleGroups MuscleGroupReader) *MuscleGroupController {
	return &MuscleGroupController{muscleGroups: muscleGroups}
}

func (m *MuscleGroupController) GetAll(c *echo.Context) error {
	muscleGroups, err := m.muscleGroups.GetAll(c.Request().Context())
	if err != nil {
		return err
	}

	response := make([]dto.MuscleGroupResponse, len(muscleGroups))
	for index, muscleGroup := range muscleGroups {
		response[index] = muscleGroupResponse(muscleGroup)
	}
	return c.JSON(http.StatusOK, response)
}

func (m *MuscleGroupController) GetByID(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "muscle group id must be a positive integer")
	}

	muscleGroup, err := m.muscleGroups.GetByID(c.Request().Context(), id)
	if err != nil {
		return muscleGroupHTTPError(err)
	}
	return c.JSON(http.StatusOK, muscleGroupResponse(muscleGroup))
}

func (m *MuscleGroupController) GetByName(c *echo.Context) error {
	muscleGroup, err := m.muscleGroups.GetByName(c.Request().Context(), c.Param("name"))
	if err != nil {
		return muscleGroupHTTPError(err)
	}
	return c.JSON(http.StatusOK, muscleGroupResponse(muscleGroup))
}

func muscleGroupResponse(muscleGroup model.MuscleGroup) dto.MuscleGroupResponse {
	return dto.MuscleGroupResponse{
		ID:          muscleGroup.ID,
		Name:        muscleGroup.Name,
		DisplayName: muscleGroup.DisplayName,
	}
}

func muscleGroupHTTPError(err error) error {
	if errors.Is(err, service.ErrMuscleGroupNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "muscle group not found")
	}
	return err
}
