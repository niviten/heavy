package model

// Excercise reflects an active row in the excercises table. The spelling is
// intentionally kept in sync with the existing database schema.
type Excercise struct {
	ID            int64
	Name          string
	MuscleGroupID int64
}

// ExcerciseQuery contains the supported collection filters.
type ExcerciseQuery struct {
	Page      int
	PageSize  int
	Search    string
	SortBy    string
	SortOrder string
}

// ExcercisePage is a page of exercises and its pagination metadata.
type ExcercisePage struct {
	Items      []Excercise
	Page       int
	PageSize   int
	TotalItems int64
	TotalPages int
}
