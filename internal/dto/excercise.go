package dto

// ExcerciseResponse is the HTTP representation of an exercise.
type ExcerciseResponse struct {
	ID            int64  `json:"excercise_id"`
	Name          string `json:"excercise_name"`
	MuscleGroupID int64  `json:"muscle_group_id"`
}

// CreateExcerciseRequest contains the fields accepted when creating an exercise.
type CreateExcerciseRequest struct {
	Name          string `json:"excercise_name"`
	MuscleGroupID int64  `json:"muscle_group_id"`
}

// UpdateExcerciseRequest contains the optional fields accepted when updating an
// exercise. Pointers distinguish omitted fields from invalid zero values.
type UpdateExcerciseRequest struct {
	Name          *string `json:"excercise_name"`
	MuscleGroupID *int64  `json:"muscle_group_id"`
}

// PaginationResponse describes the page returned by a collection endpoint.
type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// ExcerciseListResponse is the paginated exercise collection response.
type ExcerciseListResponse struct {
	Data       []ExcerciseResponse `json:"data"`
	Pagination PaginationResponse  `json:"pagination"`
}
