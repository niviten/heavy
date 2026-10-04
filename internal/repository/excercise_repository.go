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
	ErrExcerciseNotFound          = errors.New("excercise not found")
	ErrExcerciseNameConflict      = errors.New("excercise name already exists")
	ErrExcerciseMuscleGroupAbsent = errors.New("muscle group not found")
)

var excerciseSortColumns = map[string]string{
	"excercise_id":    "excercise_id",
	"excercise_name":  "excercise_name",
	"muscle_group_id": "muscle_group_id",
}

// ExcerciseRepository persists exercises in MySQL.
type ExcerciseRepository struct {
	database *sql.DB
}

func NewExcerciseRepository(database *sql.DB) *ExcerciseRepository {
	return &ExcerciseRepository{database: database}
}

func (r *ExcerciseRepository) GetAll(
	ctx context.Context,
	query model.ExcerciseQuery,
) ([]model.Excercise, int64, error) {
	return r.list(ctx, query, nil)
}

func (r *ExcerciseRepository) GetByMuscleGroup(
	ctx context.Context,
	muscleGroupID int64,
	query model.ExcerciseQuery,
) ([]model.Excercise, int64, error) {
	return r.list(ctx, query, &muscleGroupID)
}

func (r *ExcerciseRepository) GetByID(ctx context.Context, id int64) (model.Excercise, error) {
	const statement = `
		SELECT excercise_id, excercise_name, muscle_group_id
		FROM excercise
		WHERE excercise_id = ? AND is_deleted = 0`

	var excercise model.Excercise
	err := r.database.QueryRowContext(ctx, statement, id).Scan(
		&excercise.ID,
		&excercise.Name,
		&excercise.MuscleGroupID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Excercise{}, ErrExcerciseNotFound
	}
	if err != nil {
		return model.Excercise{}, fmt.Errorf("query excercise by id: %w", err)
	}
	return excercise, nil
}

func (r *ExcerciseRepository) Create(
	ctx context.Context,
	name string,
	muscleGroupID int64,
) (model.Excercise, error) {
	const statement = `
		INSERT INTO excercise (excercise_name, muscle_group_id)
		VALUES (?, ?)`

	result, err := r.database.ExecContext(ctx, statement, name, muscleGroupID)
	if err != nil {
		return model.Excercise{}, excerciseWriteError("create excercise", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.Excercise{}, fmt.Errorf("get created excercise id: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *ExcerciseRepository) Update(
	ctx context.Context,
	id int64,
	name *string,
	muscleGroupID *int64,
) (model.Excercise, error) {
	assignments := make([]string, 0, 2)
	arguments := make([]any, 0, 3)
	if name != nil {
		assignments = append(assignments, "excercise_name = ?")
		arguments = append(arguments, *name)
	}
	if muscleGroupID != nil {
		assignments = append(assignments, "muscle_group_id = ?")
		arguments = append(arguments, *muscleGroupID)
	}
	if len(assignments) == 0 {
		return r.GetByID(ctx, id)
	}
	statement := `UPDATE excercise SET ` + strings.Join(assignments, ", ") +
		` WHERE excercise_id = ? AND is_deleted = 0`
	arguments = append(arguments, id)

	_, err := r.database.ExecContext(ctx, statement, arguments...)
	if err != nil {
		return model.Excercise{}, excerciseWriteError("update excercise", err)
	}
	// Reading the row also distinguishes a no-op update from a missing or
	// already-deleted exercise.
	return r.GetByID(ctx, id)
}

func (r *ExcerciseRepository) SoftDelete(ctx context.Context, id int64) (bool, error) {
	const statement = `
		UPDATE excercise
		SET is_deleted = 1
		WHERE excercise_id = ? AND is_deleted = 0`

	result, err := r.database.ExecContext(ctx, statement, id)
	if err != nil {
		return false, fmt.Errorf("soft delete excercise: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("get soft delete result: %w", err)
	}
	return rowsAffected > 0, nil
}

func (r *ExcerciseRepository) list(
	ctx context.Context,
	query model.ExcerciseQuery,
	muscleGroupID *int64,
) ([]model.Excercise, int64, error) {
	conditions := []string{"is_deleted = 0"}
	arguments := make([]any, 0, 4)
	if muscleGroupID != nil {
		conditions = append(conditions, "muscle_group_id = ?")
		arguments = append(arguments, *muscleGroupID)
	}
	if query.Search != "" {
		conditions = append(conditions, "excercise_name LIKE ?")
		arguments = append(arguments, "%"+query.Search+"%")
	}
	whereClause := strings.Join(conditions, " AND ")

	countStatement := "SELECT COUNT(*) FROM excercise WHERE " + whereClause
	var total int64
	if err := r.database.QueryRowContext(ctx, countStatement, arguments...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count excercises: %w", err)
	}

	items := make([]model.Excercise, 0, query.PageSize)
	if total == 0 {
		return items, 0, nil
	}

	sortColumn, ok := excerciseSortColumns[query.SortBy]
	if !ok {
		sortColumn = "excercise_id"
	}
	sortOrder := "ASC"
	if strings.EqualFold(query.SortOrder, "desc") {
		sortOrder = "DESC"
	}
	statement := `
		SELECT excercise_id, excercise_name, muscle_group_id
		FROM excercise
		WHERE ` + whereClause + `
		ORDER BY ` + sortColumn + ` ` + sortOrder + `
		LIMIT ? OFFSET ?`
	listArguments := append(append([]any{}, arguments...), query.PageSize, (query.Page-1)*query.PageSize)

	rows, err := r.database.QueryContext(ctx, statement, listArguments...)
	if err != nil {
		return nil, 0, fmt.Errorf("query excercises: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var excercise model.Excercise
		if err := rows.Scan(&excercise.ID, &excercise.Name, &excercise.MuscleGroupID); err != nil {
			return nil, 0, fmt.Errorf("scan excercise: %w", err)
		}
		items = append(items, excercise)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate excercises: %w", err)
	}
	return items, total, nil
}

func excerciseWriteError(operation string, err error) error {
	var mysqlError *driver.MySQLError
	if errors.As(err, &mysqlError) {
		switch mysqlError.Number {
		case 1062:
			return ErrExcerciseNameConflict
		case 1452:
			return ErrExcerciseMuscleGroupAbsent
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}
