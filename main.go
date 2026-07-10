package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/micaroni/risk-weaver/db"
	"github.com/micaroni/risk-weaver/internal/routes"
	workload "github.com/micaroni/risk-weaver/internal/workloadhandler"
)

func init() {
	// Ignore the error because environment variables may already be
	// supplied by Docker, CI, or the shell.
	_ = godotenv.Load()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "server failure:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is not set")
	}

	// Run migrations before opening the application's connection pool.
	fmt.Println("Running migrations...")

	if err := db.RunMigrations(databaseURL); err != nil {
		return fmt.Errorf("run database migrations: %w", err)
	}

	fmt.Println("Migrations completed successfully")

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("create database connection pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	repository := workload.NewWorkloadRepository(pool)
	wlService := workload.NewWorkloadService(repository)

	fmt.Println("Database connection established")

	mux := routes.InitRoutes(wlService)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Server listening on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("start HTTP server: %w", err)
	}

	return nil
}
