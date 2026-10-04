package dto

// ExerciseResponse is the HTTP representation of an exercise.
type ExerciseResponse struct {
	ID            int64  `json:"exercise_id"`
	Name          string `json:"exercise_name"`
	MuscleGroupID int64  `json:"muscle_group_id"`
}

// CreateExerciseRequest contains the fields accepted when creating an exercise.
type CreateExerciseRequest struct {
	Name          string `json:"exercise_name"`
	MuscleGroupID int64  `json:"muscle_group_id"`
}

// UpdateExerciseRequest contains the optional fields accepted when updating an
// exercise. Pointers distinguish omitted fields from invalid zero values.
type UpdateExerciseRequest struct {
	Name          *string `json:"exercise_name"`
	MuscleGroupID *int64  `json:"muscle_group_id"`
}

// PaginationResponse describes the page returned by a collection endpoint.
type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// ExerciseListResponse is the paginated exercise collection response.
type ExerciseListResponse struct {
	Data       []ExerciseResponse `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}
