package dto

// MuscleGroupResponse is the HTTP representation of a muscle group.
type MuscleGroupResponse struct {
	ID          int64  `json:"muscle_group_id"`
	Name        string `json:"muscle_group_name"`
	DisplayName string `json:"muscle_group_display_name"`
}
