// Package dto contains data transferred across the HTTP boundary.
package dto

// HealthResponse describes application and dependency health.
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

// MessageResponse is a simple API message.
type MessageResponse struct {
	Message string `json:"message"`
}
