package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/niviten/heavy/internal/model"
)

const selectAllMuscleGroups = `
	SELECT muscle_group_id, muscle_group_name, muscle_group_display_name
	FROM muscle_group
	ORDER BY muscle_group_id`

// MuscleGroupRepository reads muscle groups from MySQL.
type MuscleGroupRepository struct {
	database *sql.DB
}

func NewMuscleGroupRepository(database *sql.DB) *MuscleGroupRepository {
	return &MuscleGroupRepository{database: database}
}

func (r *MuscleGroupRepository) GetAll(ctx context.Context) ([]model.MuscleGroup, error) {
	rows, err := r.database.QueryContext(ctx, selectAllMuscleGroups)
	if err != nil {
		return nil, fmt.Errorf("query muscle groups: %w", err)
	}
	defer rows.Close()

	muscleGroups := make([]model.MuscleGroup, 0, 11)
	for rows.Next() {
		var muscleGroup model.MuscleGroup
		if err := rows.Scan(&muscleGroup.ID, &muscleGroup.Name, &muscleGroup.DisplayName); err != nil {
			return nil, fmt.Errorf("scan muscle group: %w", err)
		}
		muscleGroups = append(muscleGroups, muscleGroup)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate muscle groups: %w", err)
	}

	return muscleGroups, nil
}
