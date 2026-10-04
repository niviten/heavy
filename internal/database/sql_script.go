package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/niviten/heavy/internal/config"
)

// SQLExecutor is implemented by sql.DB and sql.Tx.
type SQLExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// ExecuteSQLFile reads and executes a trusted SQL file as one script.
func ExecuteSQLFile(ctx context.Context, executor SQLExecutor, path string) error {
	script, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read SQL file %s: %w", path, err)
	}
	if strings.TrimSpace(string(script)) == "" {
		return fmt.Errorf("SQL file %s is empty", path)
	}
	if _, err := executor.ExecContext(ctx, string(script)); err != nil {
		return fmt.Errorf("execute SQL file %s: %w", path, err)
	}
	return nil
}

// ApplySQLFile connects to MySQL with multi-statement support and executes a
// trusted SQL file. It is intended for migration and seed commands only.
func ApplySQLFile(ctx context.Context, cfg config.DatabaseConfig, path string) (err error) {
	db, err := OpenMySQLForScripts(cfg)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close mysql: %w", closeErr))
		}
	}()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}
	return ExecuteSQLFile(ctx, db, path)
}
