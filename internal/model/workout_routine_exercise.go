package model

// WorkoutRoutineExercise is an active exercise assigned to a workout routine.
// ExerciseOrder is stored in thousandths to mirror DECIMAL(10,3) without using
// floating-point arithmetic.
type WorkoutRoutineExercise struct {
	ID               int64
	WorkoutRoutineID int64
	ExerciseID       int64
	ExerciseName     string
	MuscleGroupID    int64
	ExerciseOrder    int64
}

// WorkoutRoutineExerciseQuery contains the supported collection filters.
type WorkoutRoutineExerciseQuery struct {
	Page      int
	PageSize  int
	Search    string
	SortBy    string
	SortOrder string
}

// WorkoutRoutineExercisePage is a page of exercises assigned to a routine.
type WorkoutRoutineExercisePage struct {
	Items      []WorkoutRoutineExercise
	Page       int
	PageSize   int
	TotalItems int64
	TotalPages int
}
