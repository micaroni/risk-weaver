//go:build integration

package workload

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func setupIntegrationTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to create test database pool: %v", err)
	}

	var dbName string

	if err := pool.QueryRow(ctx, `SELECT current_database()`).Scan(&dbName); err != nil {
		pool.Close()
		t.Fatalf("failed to get current database name: %v", err)
	}

	t.Logf("integration tests using database: %s", dbName)

	if !strings.Contains(dbName, "test") {
		pool.Close()
		t.Fatalf(
			"refusing to run integration tests against non-test database %q",
			dbName,
		)
	}

	// continue setup...
	return pool
}

func TestWorkloadRepository_CreateWorkload_Success(t *testing.T) {
	pool := setupIntegrationTestDB(t)

	repository := NewWorkloadRepository(pool)

	workloadID := uuid.Must(uuid.NewV4())

	newWorkload := Workload{
		ID:          workloadID,
		Name:        "orders-api",
		Namespace:   "demo",
		Environment: "development",
		Owner:       "security team",
	}

	err := repository.CreateWorkload(context.Background(), newWorkload)
	if err != nil {
		t.Fatalf("expected CreateWorkload to succeed, got error: %v", err)
	}

	var count int

	err = pool.QueryRow(
		context.Background(),
		`SELECT COUNT(*) FROM workloads WHERE id = $1`,
		workloadID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count inserted workload: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 inserted workload, got %d", count)
	}
}

func TestWorkloadRepository_GetWorkloadByID_Success(t *testing.T) {
	pool := setupIntegrationTestDB(t)

	repository := NewWorkloadRepository(pool)

	workloadID := uuid.Must(uuid.NewV4())

	_, err := pool.Exec(
		context.Background(),
		`
			INSERT INTO workloads (
				id,
				name,
				namespace,
				environment,
				owner
			)
			VALUES ($1, $2, $3, $4, $5)
		`,
		workloadID,
		"orders-api",
		"demo",
		"development",
		"security team",
	)
	if err != nil {
		t.Fatalf("failed to insert test workload: %v", err)
	}

	got, err := repository.GetWorkloadByID(context.Background(), workloadID)
	if err != nil {
		t.Fatalf("expected GetWorkloadByID to succeed, got error: %v", err)
	}

	if got.ID != workloadID {
		t.Fatalf("expected ID %s, got %s", workloadID, got.ID)
	}

	if got.Name != "orders-api" {
		t.Fatalf("expected name %q, got %q", "orders-api", got.Name)
	}

	if got.Namespace != "demo" {
		t.Fatalf("expected namespace %q, got %q", "demo", got.Namespace)
	}

	if got.Environment != "development" {
		t.Fatalf("expected environment %q, got %q", "development", got.Environment)
	}

	if got.Owner != "security team" {
		t.Fatalf("expected owner %q, got %q", "security team", got.Owner)
	}
}

func TestWorkloadRepository_GetWorkloadByID_NotFound(t *testing.T) {
	pool := setupIntegrationTestDB(t)

	repository := NewWorkloadRepository(pool)

	missingID := uuid.Must(uuid.NewV4())

	_, err := repository.GetWorkloadByID(context.Background(), missingID)
	if err == nil {
		t.Fatal("expected error for missing workload, got nil")
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected error to wrap pgx.ErrNoRows, got %v", err)
	}
}

func TestWorkloadRepository_CreateWorkload_DuplicateID(t *testing.T) {
	pool := setupIntegrationTestDB(t)

	repository := NewWorkloadRepository(pool)

	workloadID := uuid.Must(uuid.NewV4())

	newWorkload := Workload{
		ID:          workloadID,
		Name:        "orders-api",
		Namespace:   "demo",
		Environment: "development",
		Owner:       "security team",
	}

	err := repository.CreateWorkload(context.Background(), newWorkload)
	if err != nil {
		t.Fatalf("expected first CreateWorkload to succeed, got error: %v", err)
	}

	err = repository.CreateWorkload(context.Background(), newWorkload)
	if err == nil {
		t.Fatal("expected duplicate insert to fail, got nil")
	}
}
