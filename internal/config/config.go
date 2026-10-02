package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 25
	defaultConnMaxLifetime = 5 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
)

// Config contains all application configuration.
type Config struct {
	Environment string
	Port        int
	Database    DatabaseConfig
}

// DatabaseConfig contains MySQL connection and pool settings.
type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// Address returns the HTTP listen address.
func (c Config) Address() string {
	return fmt.Sprintf(":%d", c.Port)
}

// Load reads a local .env file when present, then loads configuration from the
// environment. Existing environment variables take precedence over .env.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	return loadFrom(os.LookupEnv)
}

type lookupFunc func(string) (string, bool)

func loadFrom(lookup lookupFunc) (Config, error) {
	required := []string{
		"APP_ENV",
		"APP_PORT",
		"DB_HOST",
		"DB_PORT",
		"DB_NAME",
		"DB_USER",
		"DB_PASSWORD",
	}

	values := make(map[string]string, len(required))
	missing := make([]string, 0)
	for _, key := range required {
		value, ok := lookup(key)
		value = strings.TrimSpace(value)
		if !ok || value == "" {
			missing = append(missing, key)
			continue
		}
		values[key] = value
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	appPort, err := parsePort("APP_PORT", values["APP_PORT"])
	if err != nil {
		return Config{}, err
	}
	dbPort, err := parsePort("DB_PORT", values["DB_PORT"])
	if err != nil {
		return Config{}, err
	}

	maxOpenConns, err := optionalPositiveInt(lookup, "DB_MAX_OPEN_CONNS", defaultMaxOpenConns)
	if err != nil {
		return Config{}, err
	}
	maxIdleConns, err := optionalNonNegativeInt(lookup, "DB_MAX_IDLE_CONNS", defaultMaxIdleConns)
	if err != nil {
		return Config{}, err
	}
	if maxIdleConns > maxOpenConns {
		return Config{}, errors.New("DB_MAX_IDLE_CONNS cannot exceed DB_MAX_OPEN_CONNS")
	}

	connMaxLifetime, err := optionalDuration(lookup, "DB_CONN_MAX_LIFETIME", defaultConnMaxLifetime)
	if err != nil {
		return Config{}, err
	}
	connMaxIdleTime, err := optionalDuration(lookup, "DB_CONN_MAX_IDLE_TIME", defaultConnMaxIdleTime)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment: values["APP_ENV"],
		Port:        appPort,
		Database: DatabaseConfig{
			Host:            values["DB_HOST"],
			Port:            dbPort,
			Name:            values["DB_NAME"],
			User:            values["DB_USER"],
			Password:        values["DB_PASSWORD"],
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
			ConnMaxIdleTime: connMaxIdleTime,
		},
	}, nil
}

func parsePort(name, value string) (int, error) {
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("%s must be a number between 1 and 65535", name)
	}
	return port, nil
}

func optionalPositiveInt(lookup lookupFunc, name string, fallback int) (int, error) {
	value, ok := lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return parsed, nil
}

func optionalNonNegativeInt(lookup lookupFunc, name string, fallback int) (int, error) {
	value, ok := lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", name)
	}
	return parsed, nil
}

func optionalDuration(lookup lookupFunc, name string, fallback time.Duration) (time.Duration, error) {
	value, ok := lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", name)
	}
	return parsed, nil
}
