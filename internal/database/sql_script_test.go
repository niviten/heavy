package database

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type stubSQLExecutor struct {
	query string
	err   error
}

func (s *stubSQLExecutor) ExecContext(
	_ context.Context,
	query string,
	_ ...any,
) (sql.Result, error) {
	s.query = query
	return nil, s.err
}

func TestExecuteSQLFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "script.sql")
	want := "SELECT 1;\nSELECT 2;\n"
	if err := os.WriteFile(path, []byte(want), 0o600); err != nil {
		t.Fatalf("write test SQL file: %v", err)
	}
	executor := &stubSQLExecutor{}

	if err := ExecuteSQLFile(context.Background(), executor, path); err != nil {
		t.Fatalf("ExecuteSQLFile() error = %v", err)
	}
	if executor.query != want {
		t.Errorf("executed query = %q, want %q", executor.query, want)
	}
}

func TestExecuteSQLFileReturnsReadError(t *testing.T) {
	err := ExecuteSQLFile(
		context.Background(),
		&stubSQLExecutor{},
		filepath.Join(t.TempDir(), "missing.sql"),
	)
	if err == nil || !strings.Contains(err.Error(), "read SQL file") {
		t.Fatalf("ExecuteSQLFile() error = %v, want read error", err)
	}
}

func TestExecuteSQLFileReturnsExecutionError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "script.sql")
	if err := os.WriteFile(path, []byte("SELECT 1;"), 0o600); err != nil {
		t.Fatalf("write test SQL file: %v", err)
	}
	want := errors.New("database unavailable")

	err := ExecuteSQLFile(context.Background(), &stubSQLExecutor{err: want}, path)
	if !errors.Is(err, want) {
		t.Fatalf("ExecuteSQLFile() error = %v, want %v", err, want)
	}
}

func TestExecuteSQLFileRejectsEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.sql")
	if err := os.WriteFile(path, []byte(" \n\t"), 0o600); err != nil {
		t.Fatalf("write test SQL file: %v", err)
	}

	err := ExecuteSQLFile(context.Background(), &stubSQLExecutor{}, path)
	if err == nil || !strings.Contains(err.Error(), "is empty") {
		t.Fatalf("ExecuteSQLFile() error = %v, want empty-file error", err)
	}
}
