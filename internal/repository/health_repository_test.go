package repository

import (
	"context"
	"errors"
	"testing"
)

type stubDatabasePinger struct {
	err error
}

func (p stubDatabasePinger) PingContext(context.Context) error {
	return p.err
}

func TestHealthRepositoryPing(t *testing.T) {
	expectedError := errors.New("database unavailable")
	tests := []struct {
		name      string
		pingError error
	}{
		{name: "database available"},
		{name: "database unavailable", pingError: expectedError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewHealthRepository(stubDatabasePinger{err: tt.pingError}).Ping(context.Background())
			if !errors.Is(err, tt.pingError) {
				t.Fatalf("Ping() error = %v, want %v", err, tt.pingError)
			}
		})
	}
}
