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
	ErrWorkoutRoutineNotFound     = errors.New("workout routine not found")
	ErrWorkoutRoutineNameConflict = errors.New("workout routine name already exists")
)

var workoutRoutineSortColumns = map[string]string{
	"workout_routine_id":          "workout_routine_id",
	"workout_routine_name":        "workout_routine_name",
	"workout_routine_description": "workout_routine_description",
}

// WorkoutRoutineRepository persists workout routines in MySQL.
type WorkoutRoutineRepository struct {
	database *sql.DB
}

func NewWorkoutRoutineRepository(database *sql.DB) *WorkoutRoutineRepository {
	return &WorkoutRoutineRepository{database: database}
}

func (r *WorkoutRoutineRepository) GetAll(
	ctx context.Context,
	query model.WorkoutRoutineQuery,
) ([]model.WorkoutRoutine, int64, error) {
	arguments := make([]any, 0, 4)
	whereClause := ""
	if query.Search != "" {
		whereClause = ` WHERE workout_routine_name LIKE ? OR workout_routine_description LIKE ?`
		search := "%" + query.Search + "%"
		arguments = append(arguments, search, search)
	}

	var total int64
	countStatement := "SELECT COUNT(*) FROM workout_routines" + whereClause
	if err := r.database.QueryRowContext(ctx, countStatement, arguments...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count workout routines: %w", err)
	}

	items := make([]model.WorkoutRoutine, 0, query.PageSize)
	if total == 0 {
		return items, 0, nil
	}

	sortColumn, ok := workoutRoutineSortColumns[query.SortBy]
	if !ok {
		sortColumn = "workout_routine_id"
	}
	sortOrder := "ASC"
	if strings.EqualFold(query.SortOrder, "desc") {
		sortOrder = "DESC"
	}
	statement := `
		SELECT workout_routine_id, workout_routine_name, workout_routine_description
		FROM workout_routines` + whereClause + `
		ORDER BY ` + sortColumn + ` ` + sortOrder + `
		LIMIT ? OFFSET ?`
	listArguments := append(append([]any{}, arguments...), query.PageSize, (query.Page-1)*query.PageSize)
	rows, err := r.database.QueryContext(ctx, statement, listArguments...)
	if err != nil {
		return nil, 0, fmt.Errorf("query workout routines: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var routine model.WorkoutRoutine
		if err := rows.Scan(&routine.ID, &routine.Name, &routine.Description); err != nil {
			return nil, 0, fmt.Errorf("scan workout routine: %w", err)
		}
		items = append(items, routine)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate workout routines: %w", err)
	}
	return items, total, nil
}

func (r *WorkoutRoutineRepository) GetByID(
	ctx context.Context,
	id int64,
) (model.WorkoutRoutine, error) {
	const statement = `
		SELECT workout_routine_id, workout_routine_name, workout_routine_description
		FROM workout_routines
		WHERE workout_routine_id = ?`

	var routine model.WorkoutRoutine
	err := r.database.QueryRowContext(ctx, statement, id).Scan(
		&routine.ID,
		&routine.Name,
		&routine.Description,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.WorkoutRoutine{}, ErrWorkoutRoutineNotFound
	}
	if err != nil {
		return model.WorkoutRoutine{}, fmt.Errorf("query workout routine by id: %w", err)
	}
	return routine, nil
}

func (r *WorkoutRoutineRepository) Create(
	ctx context.Context,
	name string,
	description *string,
) (model.WorkoutRoutine, error) {
	const statement = `
		INSERT INTO workout_routines (workout_routine_name, workout_routine_description)
		VALUES (?, ?)`

	result, err := r.database.ExecContext(ctx, statement, name, description)
	if err != nil {
		return model.WorkoutRoutine{}, workoutRoutineWriteError("create workout routine", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.WorkoutRoutine{}, fmt.Errorf("get created workout routine id: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *WorkoutRoutineRepository) Update(
	ctx context.Context,
	id int64,
	name *string,
	description *string,
	updateDescription bool,
) (model.WorkoutRoutine, error) {
	assignments := make([]string, 0, 2)
	arguments := make([]any, 0, 3)
	if name != nil {
		assignments = append(assignments, "workout_routine_name = ?")
		arguments = append(arguments, *name)
	}
	if updateDescription {
		assignments = append(assignments, "workout_routine_description = ?")
		arguments = append(arguments, description)
	}
	if len(assignments) == 0 {
		return r.GetByID(ctx, id)
	}
	statement := `UPDATE workout_routines SET ` + strings.Join(assignments, ", ") +
		` WHERE workout_routine_id = ?`
	arguments = append(arguments, id)

	if _, err := r.database.ExecContext(ctx, statement, arguments...); err != nil {
		return model.WorkoutRoutine{}, workoutRoutineWriteError("update workout routine", err)
	}
	return r.GetByID(ctx, id)
}

func (r *WorkoutRoutineRepository) Delete(ctx context.Context, id int64) (bool, error) {
	const statement = `DELETE FROM workout_routines WHERE workout_routine_id = ?`

	result, err := r.database.ExecContext(ctx, statement, id)
	if err != nil {
		return false, fmt.Errorf("delete workout routine: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("get workout routine delete result: %w", err)
	}
	return rowsAffected > 0, nil
}

func workoutRoutineWriteError(operation string, err error) error {
	var mysqlError *driver.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
		return ErrWorkoutRoutineNameConflict
	}
	return fmt.Errorf("%s: %w", operation, err)
}
