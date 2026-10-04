package dto

// WorkoutRoutineResponse is the HTTP representation of a workout routine.
type WorkoutRoutineResponse struct {
	ID          int64   `json:"workout_routine_id"`
	Name        string  `json:"workout_routine_name"`
	Description *string `json:"workout_routine_description"`
}

// CreateWorkoutRoutineRequest contains the fields accepted when creating a
// workout routine.
type CreateWorkoutRoutineRequest struct {
	Name        string  `json:"workout_routine_name"`
	Description *string `json:"workout_routine_description"`
}

// UpdateWorkoutRoutineRequest contains the fields accepted when updating a
// workout routine. Pointers distinguish omitted fields from empty values.
type UpdateWorkoutRoutineRequest struct {
	Name        *string `json:"workout_routine_name"`
	Description *string `json:"workout_routine_description"`
}

// WorkoutRoutineListResponse is the paginated workout routine collection.
type WorkoutRoutineListResponse struct {
	Data       []WorkoutRoutineResponse `json:"data"`
	Pagination PaginationResponse       `json:"pagination"`
}
