package workload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/micaroni/risk-weaver/internal/utils"
)

type Workload struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Namespace   string    `json:"namespace"`
	Environment string    `json:"environment"`
	Owner       string    `json:"owner"`
}

type WorkloadService interface {
	AddNewWorkload() httprouter.Handle
	RetrieveWorkload() httprouter.Handle
}

type Repository interface {
	CreateWorkload(ctx context.Context, workload Workload) error
	GetWorkloadByID(ctx context.Context, id uuid.UUID) (Workload, error)
}

type workloadService struct {
	repository Repository
}

type WorkloadRepository struct {
	db *pgxpool.Pool
}

func NewWorkloadRepository(db *pgxpool.Pool) *WorkloadRepository {
	if db == nil {
		panic("database pool cannot be nil")
	}

	return &WorkloadRepository{
		db: db,
	}
}

func NewWorkloadService(repository Repository) WorkloadService {
	if repository == nil {
		panic("workload repository cannot be nil")
	}

	return &workloadService{
		repository: repository,
	}
}

func (wlr *WorkloadRepository) CreateWorkload(ctx context.Context, wL Workload) error {
	const query = `
		INSERT INTO workloads (
			id,
			name,
			namespace,
			environment,
			owner
		)
		VALUES ($1, $2, $3, $4, $5)
	`
	commandTag, err := wlr.db.Exec(
		ctx,
		query,
		wL.ID,
		wL.Name,
		wL.Namespace,
		wL.Environment,
		wL.Owner,
	)
	if err != nil {
		return fmt.Errorf("insert workload: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("expected one workload to be inserted")
	}

	return nil
}

func (wlr *WorkloadRepository) GetWorkloadByID(ctx context.Context, id uuid.UUID) (Workload, error) {
	const query = `
		SELECT
			id,
			name,
			namespace,
			environment,
			owner
		FROM workloads
		WHERE id = $1
	`
	var workload Workload

	err := wlr.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&workload.ID,
		&workload.Name,
		&workload.Namespace,
		&workload.Environment,
		&workload.Owner,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Workload{}, fmt.Errorf("workload not found: %w", err)
	}
	if err != nil {
		return Workload{}, fmt.Errorf("get workload by ID: %w", err)
	}

	return workload, nil
}

func (ws *workloadService) AddNewWorkload() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Printf("received new workload")

		apiStatusCode := http.StatusOK
		apiRespBody := []byte{}

		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic in AddNewWorkload: %v", rec)

				apiStatusCode = http.StatusInternalServerError
				apiRespBody = utils.GetHTTPErrMessageJSONBytes(apiStatusCode, "internal server error")
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(apiStatusCode)
			if _, err := w.Write(apiRespBody); err != nil {
				log.Printf("failed to write response: %v", err)
			}
			log.Printf("finished api call")
		}()
		defer r.Body.Close()

		var newWorkload Workload

		if err := json.NewDecoder(r.Body).Decode(&newWorkload); err != nil {
			log.Printf("failed to decode validated workload request: %v", err)

			apiStatusCode = http.StatusInternalServerError
			apiRespBody = utils.GetHTTPErrMessageJSONBytes(apiStatusCode, "failed to process workload request")
			return
		}

		newWorkload.ID = generateUUID()
		if newWorkload.ID == uuid.Nil {
			log.Printf("error returning workload ID")
			apiStatusCode = http.StatusInternalServerError
			apiRespBody = utils.GetHTTPErrMessageJSONBytes(apiStatusCode, "failed to generate workload ID")
			return
		}

		if err := ws.repository.CreateWorkload(r.Context(), newWorkload); err != nil {
			log.Printf("failed to create workload: %v", err)
			apiStatusCode = http.StatusInternalServerError
			apiRespBody = utils.GetHTTPErrMessageJSONBytes(apiStatusCode, "failed to create workload")

			return
		}

		apiStatusCode = http.StatusCreated
		apiRespBody = utils.GetHTTPErrMessageJSONBytes(apiStatusCode, "workload created with ID: "+newWorkload.ID.String())
	}
}

func (ws *workloadService) RetrieveWorkload() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Printf("retrieving workload")

		apiStatusCode := http.StatusOK
		apiRespBody := []byte{}

		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic in RetrieveWorkload: %v", rec)

				apiStatusCode = http.StatusInternalServerError
				apiRespBody = utils.GetHTTPErrMessageJSONBytes(apiStatusCode, "internal server error")
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(apiStatusCode)
			if _, err := w.Write(apiRespBody); err != nil {
				log.Printf("failed to write response: %v", err)
			}
			log.Printf("finished api call")
		}()

		idString := p.ByName("id")

		workloadID, err := uuid.FromString(idString)
		if err != nil {
			apiStatusCode = http.StatusBadRequest
			apiRespBody = utils.GetHTTPErrMessageJSONBytes(apiStatusCode, "invalid workload ID")
			return
		}

		retrievedWorkload, err := ws.repository.GetWorkloadByID(r.Context(), workloadID)
		if errors.Is(err, pgx.ErrNoRows) {
			apiStatusCode = http.StatusNotFound
			apiRespBody = utils.GetHTTPErrMessageJSONBytes(apiStatusCode, "workload not found")

			return
		}

		if err != nil {
			log.Printf("failed to get workload by ID: %v", err)
			apiStatusCode = http.StatusInternalServerError
			apiRespBody = utils.GetHTTPErrMessageJSONBytes(apiStatusCode, "failed to retrieve workload")

			return
		}

		apiRespBody, err = json.Marshal(retrievedWorkload)
		if err != nil {
			log.Printf("failed to marshal workload response: %v", err)
			apiStatusCode = http.StatusInternalServerError
			apiRespBody = utils.GetHTTPErrMessageJSONBytes(apiStatusCode, "failed to marshal workload response")

			return
		}

		apiStatusCode = http.StatusOK
	}
}

func generateUUID() uuid.UUID {
	id, err := uuid.NewV4()
	if err != nil {
		return uuid.Nil
	}
	return id
}
