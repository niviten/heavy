package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	driver "github.com/go-sql-driver/mysql"
	"github.com/niviten/heavy/internal/model"
)

var (
	ErrExerciseNotFound          = errors.New("exercise not found")
	ErrExerciseNameConflict      = errors.New("exercise name already exists")
	ErrExerciseMuscleGroupAbsent = errors.New("muscle group not found")
)

var exerciseSortColumns = map[string]string{
	"exercise_id":     "exercise_id",
	"exercise_name":   "exercise_name",
	"muscle_group_id": "muscle_group_id",
}

// ExerciseRepository persists exercises in MySQL.
type ExerciseRepository struct {
	database *sql.DB
}

func NewExerciseRepository(database *sql.DB) *ExerciseRepository {
	return &ExerciseRepository{database: database}
}

func (r *ExerciseRepository) GetAll(
	ctx context.Context,
	query model.ExerciseQuery,
) ([]model.Exercise, int64, error) {
	return r.list(ctx, query, nil)
}

func (r *ExerciseRepository) GetByMuscleGroup(
	ctx context.Context,
	muscleGroupID int64,
	query model.ExerciseQuery,
) ([]model.Exercise, int64, error) {
	return r.list(ctx, query, &muscleGroupID)
}

func (r *ExerciseRepository) GetByID(ctx context.Context, id int64) (model.Exercise, error) {
	const statement = `
		SELECT exercise_id, exercise_name, muscle_group_id
		FROM exercises
		WHERE exercise_id = ? AND is_deleted = 0`

	var exercise model.Exercise
	err := r.database.QueryRowContext(ctx, statement, id).Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.MuscleGroupID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Exercise{}, ErrExerciseNotFound
	}
	if err != nil {
		return model.Exercise{}, fmt.Errorf("query exercise by id: %w", err)
	}
	return exercise, nil
}

func (r *ExerciseRepository) Create(
	ctx context.Context,
	name string,
	muscleGroupID int64,
) (model.Exercise, error) {
	const statement = `
		INSERT INTO exercises (exercise_name, muscle_group_id)
		VALUES (?, ?)`

	result, err := r.database.ExecContext(ctx, statement, name, muscleGroupID)
	if err != nil {
		return model.Exercise{}, exerciseWriteError("create exercise", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.Exercise{}, fmt.Errorf("get created exercise id: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *ExerciseRepository) Update(
	ctx context.Context,
	id int64,
	name *string,
	muscleGroupID *int64,
) (model.Exercise, error) {
	assignments := make([]string, 0, 2)
	arguments := make([]any, 0, 3)
	if name != nil {
		assignments = append(assignments, "exercise_name = ?")
		arguments = append(arguments, *name)
	}
	if muscleGroupID != nil {
		assignments = append(assignments, "muscle_group_id = ?")
		arguments = append(arguments, *muscleGroupID)
	}
	if len(assignments) == 0 {
		return r.GetByID(ctx, id)
	}
	statement := `UPDATE exercises SET ` + strings.Join(assignments, ", ") +
		` WHERE exercise_id = ? AND is_deleted = 0`
	arguments = append(arguments, id)

	_, err := r.database.ExecContext(ctx, statement, arguments...)
	if err != nil {
		return model.Exercise{}, exerciseWriteError("update exercise", err)
	}
	// Reading the row also distinguishes a no-op update from a missing or
	// already-deleted exercise.
	return r.GetByID(ctx, id)
}

func (r *ExerciseRepository) SoftDelete(ctx context.Context, id int64) (bool, error) {
	const statement = `
		UPDATE exercises
		SET is_deleted = 1
		WHERE exercise_id = ? AND is_deleted = 0`

	result, err := r.database.ExecContext(ctx, statement, id)
	if err != nil {
		return false, fmt.Errorf("soft delete exercise: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("get soft delete result: %w", err)
	}
	return rowsAffected > 0, nil
}

func (r *ExerciseRepository) list(
	ctx context.Context,
	query model.ExerciseQuery,
	muscleGroupID *int64,
) ([]model.Exercise, int64, error) {
	conditions := []string{"is_deleted = 0"}
	arguments := make([]any, 0, 4)
	if muscleGroupID != nil {
		conditions = append(conditions, "muscle_group_id = ?")
		arguments = append(arguments, *muscleGroupID)
	}
	if query.Search != "" {
		conditions = append(conditions, "exercise_name LIKE ?")
		arguments = append(arguments, "%"+query.Search+"%")
	}
	whereClause := strings.Join(conditions, " AND ")

	countStatement := "SELECT COUNT(*) FROM exercises WHERE " + whereClause
	var total int64
	if err := r.database.QueryRowContext(ctx, countStatement, arguments...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count exercises: %w", err)
	}

	items := make([]model.Exercise, 0, query.PageSize)
	if total == 0 {
		return items, 0, nil
	}

	sortColumn, ok := exerciseSortColumns[query.SortBy]
	if !ok {
		sortColumn = "exercise_id"
	}
	sortOrder := "ASC"
	if strings.EqualFold(query.SortOrder, "desc") {
		sortOrder = "DESC"
	}
	statement := `
		SELECT exercise_id, exercise_name, muscle_group_id
		FROM exercises
		WHERE ` + whereClause + `
		ORDER BY ` + sortColumn + ` ` + sortOrder + `
		LIMIT ? OFFSET ?`
	listArguments := append(append([]any{}, arguments...), query.PageSize, (query.Page-1)*query.PageSize)

	rows, err := r.database.QueryContext(ctx, statement, listArguments...)
	if err != nil {
		return nil, 0, fmt.Errorf("query exercises: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var exercise model.Exercise
		if err := rows.Scan(&exercise.ID, &exercise.Name, &exercise.MuscleGroupID); err != nil {
			return nil, 0, fmt.Errorf("scan exercise: %w", err)
		}
		items = append(items, exercise)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate exercises: %w", err)
	}
	return items, total, nil
}

func exerciseWriteError(operation string, err error) error {
	var mysqlError *driver.MySQLError
	if errors.As(err, &mysqlError) {
		switch mysqlError.Number {
		case 1062:
			return ErrExerciseNameConflict
		case 1452:
			return ErrExerciseMuscleGroupAbsent
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}
