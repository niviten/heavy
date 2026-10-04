package model

// Exercise reflects an active row in the exercises table. The spelling is
// intentionally kept in sync with the existing database schema.
type Exercise struct {
	ID            int64
	Name          string
	MuscleGroupID int64
}

// ExerciseQuery contains the supported collection filters.
type ExerciseQuery struct {
	Page      int
	PageSize  int
	Search    string
	SortBy    string
	SortOrder string
}

// ExercisePage is a page of exercises and its pagination metadata.
type ExercisePage struct {
	Items      []Exercise
	Page       int
	PageSize   int
	TotalItems int64
	TotalPages int
}
