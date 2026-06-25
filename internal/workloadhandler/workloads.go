package workload

import (
	"errors"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	httputil "github.com/micaroni/risk-weaver/internal/httputil"
)

var (
	workloadServiceSyncOnce sync.Once
	workloadServiceObj      *workloadService
)

type Workload struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Namespace   string    `json:"namespace"`
	Environment string    `json:"environment"`
	Owner       string    `json:"owner"`
}

var WLRequest Workload

type WorkloadService interface {
	AddNewWorkload() httprouter.Handle
}

type workloadService struct {
	Workload
}

func InitWorkloadService(workload Workload) WorkloadService {
	workloadServiceSyncOnce.Do(func() {
		workloadServiceObj = &workloadService{Workload: workload}
	})

	return workloadServiceObj
}

func (ws *workloadService) AddNewWorkload() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Printf("received new workload")

		apiStatusCode := http.StatusOK
		apiRespBody := []byte{}

		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic in AddNewWorkload: %v", rec)
				apiStatusCode = http.StatusInternalServerError
				apiRespBody = httputil.GetHTTPErrMessageJSONBytes(apiStatusCode, "internal server error")
			}

			w.WriteHeader(apiStatusCode)
			if _, err := w.Write(apiRespBody); err != nil {
				log.Printf("failed to write response: %v", err)
			}
			log.Printf("finished api call")
		}()

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		defer r.Body.Close()

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("failed to read request body: %v", err)
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				apiStatusCode = http.StatusRequestEntityTooLarge
				apiRespBody = httputil.GetHTTPErrMessageJSONBytes(apiStatusCode, "request body too large")
			} else {
				apiStatusCode = http.StatusInternalServerError
				apiRespBody = httputil.GetHTTPErrMessageJSONBytes(apiStatusCode, "failed to read request body")
			}
			return
		}

		if len(reqBody) == 0 {
			log.Printf("received empty request body")
			apiStatusCode = http.StatusBadRequest
			apiRespBody = httputil.GetHTTPErrMessageJSONBytes(apiStatusCode, "empty request body")
			return
		}

		var newWorkload Workload
		err = httputil.UnmarshalJSON(reqBody, &newWorkload)
		if err != nil {
			log.Printf("error unmarshalling request")
			apiStatusCode = http.StatusBadRequest
			apiRespBody = httputil.GetHTTPErrMessageJSONBytes(apiStatusCode, "error unmarshalling request")
			return
		}

		newWorkload.ID = generateUUID()
		if newWorkload.ID == uuid.Nil {
			log.Printf("error returning workload ID")
			apiStatusCode = http.StatusInternalServerError
			apiRespBody = httputil.GetHTTPErrMessageJSONBytes(apiStatusCode, "error returning workload ID")
			return
		} else {
			stringUUID := newWorkload.ID.String()
			apiRespBody = httputil.GetHTTPErrMessageJSONBytes(apiStatusCode, "Workload ID: "+stringUUID)
		}
	}
}

func generateUUID() uuid.UUID {
	id, err := uuid.NewV4()
	if err != nil {
		return uuid.Nil
	}

	addUUIDtoWorkload(id)

	return id
}

func addUUIDtoWorkload(uuid uuid.UUID) error {
	return nil
}
