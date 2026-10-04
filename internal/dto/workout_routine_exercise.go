package dto

// AddWorkoutRoutineExerciseRequest identifies the exercise to append.
type AddWorkoutRoutineExerciseRequest struct {
	ExerciseID int64 `json:"exercise_id"`
}

// BulkAddWorkoutRoutineExercisesRequest identifies exercises to append in order.
type BulkAddWorkoutRoutineExercisesRequest struct {
	ExerciseIDs []int64 `json:"exercise_ids"`
}

// ReorderWorkoutRoutineExerciseRequest gives the exercise's one-based final position.
type ReorderWorkoutRoutineExerciseRequest struct {
	DestinationPosition int `json:"destination_position"`
}

// WorkoutRoutineExerciseResponse is an exercise's representation within a routine.
type WorkoutRoutineExerciseResponse struct {
	ID               int64  `json:"workout_routine_exercise_id"`
	WorkoutRoutineID int64  `json:"workout_routine_id"`
	ExerciseID       int64  `json:"exercise_id"`
	ExerciseName     string `json:"exercise_name"`
	MuscleGroupID    int64  `json:"muscle_group_id"`
	ExerciseOrder    string `json:"exercise_order"`
}

// WorkoutRoutineExerciseListResponse is the paginated routine exercise collection.
type WorkoutRoutineExerciseListResponse struct {
	Data       []WorkoutRoutineExerciseResponse `json:"data"`
	Pagination PaginationResponse               `json:"pagination"`
}

// BulkAddWorkoutRoutineExercisesResponse contains the newly assigned exercises.
type BulkAddWorkoutRoutineExercisesResponse struct {
	Data []WorkoutRoutineExerciseResponse `json:"data"`
}
