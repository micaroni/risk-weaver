package workload

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
)

type fakeRepository struct {
	createWorkloadFunc func(
		ctx context.Context,
		workload Workload,
	) error

	getWorkloadByIDFunc func(
		ctx context.Context,
		id uuid.UUID,
	) (Workload, error)
}

func TestGenerateUUID_Success(t *testing.T) {
	id := generateUUID()

	if id == uuid.Nil {
		t.Fatal("expected generated UUID to be non-nil")
	}

	if _, err := uuid.FromString(id.String()); err != nil {
		t.Fatalf("expected valid UUID string, got error: %v", err)
	}
}

func TestNewWorkloadRepository_PanicsWithNilDB(t *testing.T) {
	defer func() {
		rec := recover()
		if rec == nil {
			t.Fatal("expected panic when db is nil")
		}

		expected := "database pool cannot be nil"
		if rec != expected {
			t.Fatalf("expected panic %q, got %q", expected, rec)
		}
	}()

	NewWorkloadRepository(nil)
}

func TestNewWorkloadService_Success(t *testing.T) {
	repository := &WorkloadRepository{}

	service := NewWorkloadService(repository)

	if service == nil {
		t.Fatal("expected workload service to be non-nil")
	}
}

func TestNewWorkloadService_PanicsWithNilRepository(t *testing.T) {
	defer func() {
		rec := recover()
		if rec == nil {
			t.Fatal("expected panic when repository is nil")
		}

		expected := "workload repository cannot be nil"
		if rec != expected {
			t.Fatalf("expected panic %q, got %q", expected, rec)
		}
	}()

	NewWorkloadService(nil)
}

func (f *fakeRepository) CreateWorkload(
	ctx context.Context,
	workload Workload,
) error {
	if f.createWorkloadFunc == nil {
		return nil
	}

	return f.createWorkloadFunc(ctx, workload)
}

func (f *fakeRepository) GetWorkloadByID(
	ctx context.Context,
	id uuid.UUID,
) (Workload, error) {
	if f.getWorkloadByIDFunc == nil {
		return Workload{}, nil
	}

	return f.getWorkloadByIDFunc(ctx, id)
}
