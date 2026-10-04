package model

// WorkoutRoutine reflects a row in the workout_routines table.
type WorkoutRoutine struct {
	ID          int64
	Name        string
	Description *string
}

// WorkoutRoutineQuery contains the supported collection filters.
type WorkoutRoutineQuery struct {
	Page      int
	PageSize  int
	Search    string
	SortBy    string
	SortOrder string
}

// WorkoutRoutinePage is a page of workout routines and its pagination metadata.
type WorkoutRoutinePage struct {
	Items      []WorkoutRoutine
	Page       int
	PageSize   int
	TotalItems int64
	TotalPages int
}
