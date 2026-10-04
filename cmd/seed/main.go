package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/niviten/heavy/internal/config"
	"github.com/niviten/heavy/internal/database"
)

const seedPath = "sql/seed.sql"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seeding failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("applied %s\n", seedPath)
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return database.ApplySQLFile(ctx, cfg.Database, seedPath)
}
