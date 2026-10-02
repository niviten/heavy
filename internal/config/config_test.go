package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadFrom(t *testing.T) {
	values := validValues()
	lookup := mapLookup(values)

	cfg, err := loadFrom(lookup)
	if err != nil {
		t.Fatalf("loadFrom() error = %v", err)
	}

	if cfg.Environment != "test" {
		t.Errorf("Environment = %q, want test", cfg.Environment)
	}
	if cfg.Address() != ":8080" {
		t.Errorf("Address() = %q, want :8080", cfg.Address())
	}
	if cfg.Database.Port != 3306 {
		t.Errorf("Database.Port = %d, want 3306", cfg.Database.Port)
	}
	if cfg.Database.MaxOpenConns != defaultMaxOpenConns {
		t.Errorf("MaxOpenConns = %d, want %d", cfg.Database.MaxOpenConns, defaultMaxOpenConns)
	}
	if cfg.Database.ConnMaxLifetime != defaultConnMaxLifetime {
		t.Errorf("ConnMaxLifetime = %s, want %s", cfg.Database.ConnMaxLifetime, defaultConnMaxLifetime)
	}
}

func TestLoadFromReportsMissingVariables(t *testing.T) {
	values := validValues()
	delete(values, "DB_PASSWORD")

	_, err := loadFrom(mapLookup(values))
	if err == nil || !strings.Contains(err.Error(), "DB_PASSWORD") {
		t.Fatalf("loadFrom() error = %v, want missing DB_PASSWORD", err)
	}
}

func TestLoadFromRejectsInvalidPort(t *testing.T) {
	values := validValues()
	values["APP_PORT"] = "70000"

	_, err := loadFrom(mapLookup(values))
	if err == nil || !strings.Contains(err.Error(), "APP_PORT") {
		t.Fatalf("loadFrom() error = %v, want invalid APP_PORT", err)
	}
}

func TestLoadFromReadsPoolSettings(t *testing.T) {
	values := validValues()
	values["DB_MAX_OPEN_CONNS"] = "12"
	values["DB_MAX_IDLE_CONNS"] = "4"
	values["DB_CONN_MAX_LIFETIME"] = "3m"
	values["DB_CONN_MAX_IDLE_TIME"] = "45s"

	cfg, err := loadFrom(mapLookup(values))
	if err != nil {
		t.Fatalf("loadFrom() error = %v", err)
	}

	if cfg.Database.MaxOpenConns != 12 || cfg.Database.MaxIdleConns != 4 {
		t.Errorf("pool sizes = (%d, %d), want (12, 4)", cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns)
	}
	if cfg.Database.ConnMaxLifetime != 3*time.Minute || cfg.Database.ConnMaxIdleTime != 45*time.Second {
		t.Errorf("pool durations = (%s, %s), want (3m, 45s)", cfg.Database.ConnMaxLifetime, cfg.Database.ConnMaxIdleTime)
	}
}

func validValues() map[string]string {
	return map[string]string{
		"APP_ENV":     "test",
		"APP_PORT":    "8080",
		"DB_HOST":     "127.0.0.1",
		"DB_PORT":     "3306",
		"DB_NAME":     "heavy_test",
		"DB_USER":     "heavy",
		"DB_PASSWORD": "secret",
	}
}

func mapLookup(values map[string]string) lookupFunc {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
