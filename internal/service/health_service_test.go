package service

import (
	"context"
	"errors"
	"testing"
)

type stubHealthRepository struct {
	err error
}

func (r stubHealthRepository) Ping(context.Context) error {
	return r.err
}

func TestHealthServiceCheck(t *testing.T) {
	tests := []struct {
		name       string
		pingError  error
		wantStatus string
		wantDB     string
		wantError  bool
	}{
		{name: "healthy", wantStatus: "ok", wantDB: "up"},
		{name: "database unavailable", pingError: errors.New("connection refused"), wantStatus: "unhealthy", wantDB: "down", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewHealthService(stubHealthRepository{err: tt.pingError}).Check(context.Background())
			if (err != nil) != tt.wantError {
				t.Fatalf("Check() error = %v, wantError %v", err, tt.wantError)
			}
			if result.Status != tt.wantStatus || result.Database != tt.wantDB {
				t.Errorf("Check() = %+v, want status=%q database=%q", result, tt.wantStatus, tt.wantDB)
			}
		})
	}
}
