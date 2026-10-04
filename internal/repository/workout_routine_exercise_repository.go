package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	driver "github.com/go-sql-driver/mysql"
	"github.com/niviten/heavy/internal/model"
)

const (
	exerciseOrderScale int64 = 1000
	exerciseOrderStep  int64 = 1000
	maxExerciseOrder   int64 = 9_999_999_999
)

var (
	ErrWorkoutRoutineExerciseNotFound = errors.New("exercise is not in workout routine")
	ErrWorkoutRoutineExerciseConflict = errors.New("exercise is already in workout routine")
	ErrInvalidExercisePosition        = errors.New("destination position is outside the workout routine")
)

var workoutRoutineExerciseSortColumns = map[string]string{
	"exercise_order":  "wre.exercise_order",
	"exercise_id":     "e.exercise_id",
	"exercise_name":   "e.exercise_name",
	"muscle_group_id": "e.muscle_group_id",
}

// WorkoutRoutineExerciseRepository persists exercise membership and ordering.
type WorkoutRoutineExerciseRepository struct {
	database *sql.DB
}

func NewWorkoutRoutineExerciseRepository(database *sql.DB) *WorkoutRoutineExerciseRepository {
	return &WorkoutRoutineExerciseRepository{database: database}
}

func (r *WorkoutRoutineExerciseRepository) GetAll(
	ctx context.Context,
	workoutRoutineID int64,
	query model.WorkoutRoutineExerciseQuery,
) ([]model.WorkoutRoutineExercise, int64, error) {
	if err := workoutRoutineExists(ctx, r.database, workoutRoutineID); err != nil {
		return nil, 0, err
	}

	arguments := []any{workoutRoutineID}
	whereClause := "wre.workout_routine_id = ? AND e.is_deleted = 0"
	if query.Search != "" {
		whereClause += " AND e.exercise_name LIKE ?"
		arguments = append(arguments, "%"+query.Search+"%")
	}

	countStatement := `
		SELECT COUNT(*)
		FROM workout_routine_exercises wre
		JOIN exercises e ON e.exercise_id = wre.exercise_id
		WHERE ` + whereClause
	var total int64
	if err := r.database.QueryRowContext(ctx, countStatement, arguments...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count workout routine exercises: %w", err)
	}

	items := make([]model.WorkoutRoutineExercise, 0, query.PageSize)
	if total == 0 {
		return items, 0, nil
	}

	sortColumn, ok := workoutRoutineExerciseSortColumns[query.SortBy]
	if !ok {
		sortColumn = "wre.exercise_order"
	}
	sortOrder := "ASC"
	if strings.EqualFold(query.SortOrder, "desc") {
		sortOrder = "DESC"
	}
	statement := `
		SELECT wre.id, wre.workout_routine_id, e.exercise_id, e.exercise_name,
		       e.muscle_group_id, wre.exercise_order
		FROM workout_routine_exercises wre
		JOIN exercises e ON e.exercise_id = wre.exercise_id
		WHERE ` + whereClause + `
		ORDER BY ` + sortColumn + ` ` + sortOrder + `, wre.id ASC
		LIMIT ? OFFSET ?`
	listArguments := append(append([]any{}, arguments...), query.PageSize, (query.Page-1)*query.PageSize)
	rows, err := r.database.QueryContext(ctx, statement, listArguments...)
	if err != nil {
		return nil, 0, fmt.Errorf("query workout routine exercises: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		item, err := scanWorkoutRoutineExercise(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate workout routine exercises: %w", err)
	}
	return items, total, nil
}

func (r *WorkoutRoutineExerciseRepository) Create(
	ctx context.Context,
	workoutRoutineID int64,
	exerciseID int64,
) (model.WorkoutRoutineExercise, error) {
	items, err := r.createMany(ctx, workoutRoutineID, []int64{exerciseID})
	if err != nil {
		return model.WorkoutRoutineExercise{}, err
	}
	return items[0], nil
}

func (r *WorkoutRoutineExerciseRepository) CreateMany(
	ctx context.Context,
	workoutRoutineID int64,
	exerciseIDs []int64,
) ([]model.WorkoutRoutineExercise, error) {
	return r.createMany(ctx, workoutRoutineID, exerciseIDs)
}

func (r *WorkoutRoutineExerciseRepository) createMany(
	ctx context.Context,
	workoutRoutineID int64,
	exerciseIDs []int64,
) (items []model.WorkoutRoutineExercise, err error) {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin add workout routine exercises: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = lockWorkoutRoutine(ctx, tx, workoutRoutineID); err != nil {
		return nil, err
	}
	lockIDs := append([]int64(nil), exerciseIDs...)
	sort.Slice(lockIDs, func(i, j int) bool { return lockIDs[i] < lockIDs[j] })
	for _, exerciseID := range lockIDs {
		if err = lockActiveExercise(ctx, tx, exerciseID); err != nil {
			return nil, err
		}
	}

	var rawMaximum string
	if err = tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(MAX(exercise_order), 0)
		 FROM workout_routine_exercises
		 WHERE workout_routine_id = ?`,
		workoutRoutineID,
	).Scan(&rawMaximum); err != nil {
		return nil, fmt.Errorf("query maximum workout routine exercise order: %w", err)
	}
	maximum, parseErr := parseExerciseOrder(rawMaximum)
	if parseErr != nil {
		return nil, parseErr
	}
	exerciseCount := int64(len(exerciseIDs))
	if exerciseCount > maxExerciseOrder/exerciseOrderStep {
		return nil, fmt.Errorf("too many exercises to add to workout routine")
	}
	requiredOrderSpace := exerciseCount * exerciseOrderStep
	if maximum > maxExerciseOrder-requiredOrderSpace {
		var rows []workoutRoutineExerciseOrder
		rows, err = loadWorkoutRoutineExerciseOrders(ctx, tx, workoutRoutineID, false)
		if err != nil {
			return nil, err
		}
		if err = normalizeWorkoutRoutineExerciseOrders(ctx, tx, rows); err != nil {
			return nil, err
		}
		maximum = int64(len(rows)) * exerciseOrderStep
	}
	if maximum > maxExerciseOrder-requiredOrderSpace {
		return nil, fmt.Errorf("workout routine has no available exercise order")
	}

	items = make([]model.WorkoutRoutineExercise, 0, len(exerciseIDs))
	for index, exerciseID := range exerciseIDs {
		newOrder := maximum + int64(index+1)*exerciseOrderStep
		result, execErr := tx.ExecContext(
			ctx,
			`INSERT INTO workout_routine_exercises
			 (workout_routine_id, exercise_id, exercise_order)
			 VALUES (?, ?, ?)`,
			workoutRoutineID,
			exerciseID,
			formatExerciseOrder(newOrder),
		)
		if execErr != nil {
			return nil, workoutRoutineExerciseWriteError("add exercise to workout routine", execErr)
		}
		id, idErr := result.LastInsertId()
		if idErr != nil {
			return nil, fmt.Errorf("get workout routine exercise id: %w", idErr)
		}
		item, getErr := getWorkoutRoutineExerciseByID(ctx, tx, id)
		if getErr != nil {
			return nil, getErr
		}
		items = append(items, item)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit add workout routine exercises: %w", err)
	}
	return items, nil
}

func (r *WorkoutRoutineExerciseRepository) Delete(
	ctx context.Context,
	workoutRoutineID int64,
	exerciseID int64,
) (deleted bool, err error) {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin remove workout routine exercise: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = lockWorkoutRoutine(ctx, tx, workoutRoutineID); err != nil {
		return false, err
	}
	result, err := tx.ExecContext(
		ctx,
		`DELETE FROM workout_routine_exercises
		 WHERE workout_routine_id = ? AND exercise_id = ?`,
		workoutRoutineID,
		exerciseID,
	)
	if err != nil {
		return false, fmt.Errorf("remove exercise from workout routine: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("get workout routine exercise delete result: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("commit remove workout routine exercise: %w", err)
	}
	return rowsAffected > 0, nil
}

func (r *WorkoutRoutineExerciseRepository) Reorder(
	ctx context.Context,
	workoutRoutineID int64,
	exerciseID int64,
	destinationPosition int,
) (item model.WorkoutRoutineExercise, err error) {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return item, fmt.Errorf("begin reorder workout routine exercise: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = lockWorkoutRoutine(ctx, tx, workoutRoutineID); err != nil {
		return item, err
	}
	orders, err := loadWorkoutRoutineExerciseOrders(ctx, tx, workoutRoutineID, true)
	if err != nil {
		return item, err
	}

	currentIndex := -1
	for index := range orders {
		if orders[index].ExerciseID == exerciseID {
			currentIndex = index
			break
		}
	}
	if currentIndex == -1 {
		return item, ErrWorkoutRoutineExerciseNotFound
	}
	if destinationPosition < 1 || destinationPosition > len(orders) {
		return item, ErrInvalidExercisePosition
	}
	if currentIndex == destinationPosition-1 {
		item, err = getWorkoutRoutineExerciseByID(ctx, tx, orders[currentIndex].ID)
		if err != nil {
			return item, err
		}
		if err = tx.Commit(); err != nil {
			return model.WorkoutRoutineExercise{}, fmt.Errorf("commit reorder workout routine exercise: %w", err)
		}
		return item, nil
	}

	moving := orders[currentIndex]
	remaining := append([]workoutRoutineExerciseOrder{}, orders[:currentIndex]...)
	remaining = append(remaining, orders[currentIndex+1:]...)
	newOrder, available := calculateWorkoutRoutineExerciseOrder(remaining, destinationPosition-1)
	if !available {
		if err = normalizeWorkoutRoutineExerciseOrders(ctx, tx, orders); err != nil {
			return item, err
		}
		for index := range orders {
			orders[index].Order = int64(index+1) * exerciseOrderStep
		}
		remaining = append([]workoutRoutineExerciseOrder{}, orders[:currentIndex]...)
		remaining = append(remaining, orders[currentIndex+1:]...)
		newOrder, available = calculateWorkoutRoutineExerciseOrder(remaining, destinationPosition-1)
		if !available {
			return item, fmt.Errorf("workout routine has no available exercise order")
		}
	}

	if _, err = tx.ExecContext(
		ctx,
		`UPDATE workout_routine_exercises SET exercise_order = ? WHERE id = ?`,
		formatExerciseOrder(newOrder),
		moving.ID,
	); err != nil {
		return item, fmt.Errorf("update workout routine exercise order: %w", err)
	}
	item, err = getWorkoutRoutineExerciseByID(ctx, tx, moving.ID)
	if err != nil {
		return item, err
	}
	if err = tx.Commit(); err != nil {
		return model.WorkoutRoutineExercise{}, fmt.Errorf("commit reorder workout routine exercise: %w", err)
	}
	return item, nil
}

type rowScanner interface {
	Scan(...any) error
}

type contextQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func workoutRoutineExists(ctx context.Context, queryer contextQueryer, id int64) error {
	var found int64
	err := queryer.QueryRowContext(
		ctx,
		`SELECT workout_routine_id FROM workout_routines WHERE workout_routine_id = ?`,
		id,
	).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrWorkoutRoutineNotFound
	}
	if err != nil {
		return fmt.Errorf("query workout routine: %w", err)
	}
	return nil
}

func lockWorkoutRoutine(ctx context.Context, tx *sql.Tx, id int64) error {
	var found int64
	err := tx.QueryRowContext(
		ctx,
		`SELECT workout_routine_id
		 FROM workout_routines
		 WHERE workout_routine_id = ?
		 FOR UPDATE`,
		id,
	).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrWorkoutRoutineNotFound
	}
	if err != nil {
		return fmt.Errorf("lock workout routine: %w", err)
	}
	return nil
}

func lockActiveExercise(ctx context.Context, tx *sql.Tx, id int64) error {
	var found int64
	err := tx.QueryRowContext(
		ctx,
		`SELECT exercise_id
		 FROM exercises
		 WHERE exercise_id = ? AND is_deleted = 0
		 FOR UPDATE`,
		id,
	).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrExerciseNotFound
	}
	if err != nil {
		return fmt.Errorf("lock exercise: %w", err)
	}
	return nil
}

func getWorkoutRoutineExerciseByID(
	ctx context.Context,
	tx *sql.Tx,
	id int64,
) (model.WorkoutRoutineExercise, error) {
	row := tx.QueryRowContext(
		ctx,
		`SELECT wre.id, wre.workout_routine_id, e.exercise_id, e.exercise_name,
		        e.muscle_group_id, wre.exercise_order
		 FROM workout_routine_exercises wre
		 JOIN exercises e ON e.exercise_id = wre.exercise_id
		 WHERE wre.id = ? AND e.is_deleted = 0`,
		id,
	)
	item, err := scanWorkoutRoutineExercise(row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.WorkoutRoutineExercise{}, ErrWorkoutRoutineExerciseNotFound
	}
	return item, err
}

func scanWorkoutRoutineExercise(scanner rowScanner) (model.WorkoutRoutineExercise, error) {
	var item model.WorkoutRoutineExercise
	var rawOrder string
	err := scanner.Scan(
		&item.ID,
		&item.WorkoutRoutineID,
		&item.ExerciseID,
		&item.ExerciseName,
		&item.MuscleGroupID,
		&rawOrder,
	)
	if err != nil {
		return item, err
	}
	item.ExerciseOrder, err = parseExerciseOrder(rawOrder)
	if err != nil {
		return model.WorkoutRoutineExercise{}, err
	}
	return item, nil
}

type workoutRoutineExerciseOrder struct {
	ID         int64
	ExerciseID int64
	Order      int64
}

func loadWorkoutRoutineExerciseOrders(
	ctx context.Context,
	tx *sql.Tx,
	workoutRoutineID int64,
	activeOnly bool,
) ([]workoutRoutineExerciseOrder, error) {
	statement := `
		SELECT wre.id, wre.exercise_id, wre.exercise_order
		FROM workout_routine_exercises wre`
	if activeOnly {
		statement += ` JOIN exercises e ON e.exercise_id = wre.exercise_id AND e.is_deleted = 0`
	}
	statement += `
		WHERE wre.workout_routine_id = ?
		ORDER BY wre.exercise_order ASC, wre.id ASC
		FOR UPDATE`
	rows, err := tx.QueryContext(ctx, statement, workoutRoutineID)
	if err != nil {
		return nil, fmt.Errorf("query workout routine exercise orders: %w", err)
	}
	defer rows.Close()

	result := make([]workoutRoutineExerciseOrder, 0)
	for rows.Next() {
		var order workoutRoutineExerciseOrder
		var rawOrder string
		if err := rows.Scan(&order.ID, &order.ExerciseID, &rawOrder); err != nil {
			return nil, fmt.Errorf("scan workout routine exercise order: %w", err)
		}
		order.Order, err = parseExerciseOrder(rawOrder)
		if err != nil {
			return nil, err
		}
		result = append(result, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workout routine exercise orders: %w", err)
	}
	return result, nil
}

func calculateWorkoutRoutineExerciseOrder(
	remaining []workoutRoutineExerciseOrder,
	insertionIndex int,
) (int64, bool) {
	if insertionIndex < 0 || insertionIndex > len(remaining) {
		return 0, false
	}
	if len(remaining) == 0 {
		return exerciseOrderStep, true
	}
	if insertionIndex == 0 {
		candidate := remaining[0].Order / 2
		return candidate, candidate > 0 && candidate < remaining[0].Order
	}
	if insertionIndex == len(remaining) {
		previous := remaining[len(remaining)-1].Order
		candidate := previous + exerciseOrderStep
		return candidate, candidate > previous && candidate <= maxExerciseOrder
	}
	previous := remaining[insertionIndex-1].Order
	next := remaining[insertionIndex].Order
	candidate := previous + (next-previous)/2
	return candidate, candidate > previous && candidate < next
}

func normalizeWorkoutRoutineExerciseOrders(
	ctx context.Context,
	tx *sql.Tx,
	orders []workoutRoutineExerciseOrder,
) error {
	if int64(len(orders)) > maxExerciseOrder/exerciseOrderStep {
		return fmt.Errorf("workout routine has too many exercises to normalize order")
	}
	for index, order := range orders {
		newOrder := int64(index+1) * exerciseOrderStep
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE workout_routine_exercises SET exercise_order = ? WHERE id = ?`,
			formatExerciseOrder(newOrder),
			order.ID,
		); err != nil {
			return fmt.Errorf("normalize workout routine exercise order: %w", err)
		}
	}
	return nil
}

func parseExerciseOrder(value string) (int64, error) {
	negative := strings.HasPrefix(value, "-")
	value = strings.TrimPrefix(value, "-")
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("parse workout routine exercise order %q", value)
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse workout routine exercise order %q: %w", value, err)
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if len(fraction) > 3 {
		return 0, fmt.Errorf("parse workout routine exercise order %q", value)
	}
	fraction += strings.Repeat("0", 3-len(fraction))
	fractionValue := int64(0)
	if fraction != "" {
		fractionValue, err = strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse workout routine exercise order %q: %w", value, err)
		}
	}
	result := whole*exerciseOrderScale + fractionValue
	if negative {
		result = -result
	}
	return result, nil
}

func formatExerciseOrder(value int64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	return fmt.Sprintf("%s%d.%03d", sign, value/exerciseOrderScale, value%exerciseOrderScale)
}

func workoutRoutineExerciseWriteError(operation string, err error) error {
	var mysqlError *driver.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
		return ErrWorkoutRoutineExerciseConflict
	}
	return fmt.Errorf("%s: %w", operation, err)
}
