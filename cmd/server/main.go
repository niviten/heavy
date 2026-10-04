package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/niviten/heavy/internal/config"
	"github.com/niviten/heavy/internal/controller"
	"github.com/niviten/heavy/internal/database"
	"github.com/niviten/heavy/internal/repository"
	"github.com/niviten/heavy/internal/routes"
	"github.com/niviten/heavy/internal/service"
)

const (
	readHeaderTimeout  = 5 * time.Second
	readTimeout        = 15 * time.Second
	writeTimeout       = 15 * time.Second
	idleTimeout        = 60 * time.Second
	shutdownTimeout    = 10 * time.Second
	startupPingTimeout = 5 * time.Second
	maxRequestBody     = 1 << 20 // 1 MiB
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		logger.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.OpenMySQL(cfg.Database)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("close mysql", "error", err)
		}
	}()

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.BodyLimit(maxRequestBody))

	healthRepository := repository.NewHealthRepository(db)
	pingCtx, cancelPing := context.WithTimeout(rootCtx, startupPingTimeout)
	pingErr := healthRepository.Ping(pingCtx)
	cancelPing()
	if pingErr != nil {
		return errors.Join(errors.New("ping mysql"), pingErr)
	}

	healthService := service.NewHealthService(healthRepository)
	healthController := controller.NewHealthController(healthService)
	muscleGroupRepository := repository.NewMuscleGroupRepository(db)
	muscleGroupService := service.NewMuscleGroupService(muscleGroupRepository)
	muscleGroupController := controller.NewMuscleGroupController(muscleGroupService)
	exerciseRepository := repository.NewExerciseRepository(db)
	exerciseService := service.NewExerciseService(exerciseRepository, muscleGroupService)
	exerciseController := controller.NewExerciseController(exerciseService)
	workoutRoutineRepository := repository.NewWorkoutRoutineRepository(db)
	workoutRoutineService := service.NewWorkoutRoutineService(workoutRoutineRepository)
	workoutRoutineController := controller.NewWorkoutRoutineController(workoutRoutineService)
	workoutRoutineExerciseRepository := repository.NewWorkoutRoutineExerciseRepository(db)
	workoutRoutineExerciseService := service.NewWorkoutRoutineExerciseService(
		workoutRoutineExerciseRepository,
	)
	workoutRoutineExerciseController := controller.NewWorkoutRoutineExerciseController(
		workoutRoutineExerciseService,
	)
	routes.Register(
		e,
		healthController,
		muscleGroupController,
		exerciseController,
		workoutRoutineController,
		workoutRoutineExerciseController,
	)

	server := &http.Server{
		Addr:              cfg.Address(),
		Handler:           e,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("server starting", "address", cfg.Address(), "environment", cfg.Environment)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-rootCtx.Done():
		slog.Info("server shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return errors.Join(errors.New("graceful shutdown failed"), err)
	}

	return nil
}
