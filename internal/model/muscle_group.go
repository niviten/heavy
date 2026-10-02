// Package model contains types that reflect database tables.
package model

// MuscleGroup reflects a row in the muscle_group table.
type MuscleGroup struct {
	ID          int64
	Name        string
	DisplayName string
}
