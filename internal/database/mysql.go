package database

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"github.com/niviten/heavy/internal/config"
)

// OpenMySQL creates and configures a MySQL connection pool. Database queries
// and health checks belong in the repository layer.
func OpenMySQL(cfg config.DatabaseConfig) (*sql.DB, error) {
	driverConfig := driver.NewConfig()
	driverConfig.User = cfg.User
	driverConfig.Passwd = cfg.Password
	driverConfig.Net = "tcp"
	driverConfig.Addr = net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	driverConfig.DBName = cfg.Name
	driverConfig.ParseTime = true
	driverConfig.Loc = time.UTC
	driverConfig.Params = map[string]string{"charset": "utf8mb4"}

	connector, err := driver.NewConnector(driverConfig)
	if err != nil {
		return nil, fmt.Errorf("create mysql connector: %w", err)
	}

	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return db, nil
}
