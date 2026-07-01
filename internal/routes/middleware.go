package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/micaroni/risk-weaver/internal/utils"
)

type CreateWorkloadRequest struct {
	Name        string `json:"name"`
	Namespace   string `json: "namespace"`
	Environment string `json:"environment"`
	Owner       string `json:"owner"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Middleware func(httprouter.Handle) httprouter.Handle

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	responseBody := utils.GetHTTPErrMessageJSONBytes(statusCode, message)

	w.Header().Set("Content-Type", "appliation/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(responseBody); err != nil {
		log.Printf("failed to write error response: %v", err)
	}
}

func validateNewWorkload(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Println(r.URL.Path, "validating workload request")

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		defer r.Body.Close()

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("failed to read request body: %v", err)

			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				writeJSONError(w, http.StatusRequestEntityTooLarge, "request body too large")
			} else {
				writeJSONError(w, http.StatusInternalServerError, "failed to read request body")
			}
			return
		}

		if len(reqBody) == 0 {
			writeJSONError(w, http.StatusBadRequest, "empty request body")
			return
		}

		var request CreateWorkloadRequest
		if err := json.Unmarshal(reqBody, &request); err != nil {
			log.Printf("error unmarshalling request: %v", err)

			writeJSONError(w, http.StatusBadRequest, "invalid JSON request")
			return
		}

		switch {
		case request.Name == "":
			writeJSONError(w, http.StatusBadRequest, "name is required")
			return

		case request.Namespace == "":
			writeJSONError(w, http.StatusBadRequest, "namespace is required")
			return

		case request.Environment == "":
			writeJSONError(w, http.StatusBadRequest, "environment is required")
			return

		case request.Owner == "":
			writeJSONError(w, http.StatusBadRequest, "owner is required")
			return
		}

		// Reading r.Body consumes it. Restore it so AddNewWorkload
		// can read and unmarshal the same request body.
		r.Body = io.NopCloser(bytes.NewReader(reqBody))

		next(w, r, p)
	}
}

func contentTypeMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Println(r.URL.Path, "validating Content-Type")

		contentType := r.Header.Get("Content-Type")

		if contentType != "application/json" {
			http.Error(w, "Invalid Content-Type: application/json expected", http.StatusUnsupportedMediaType)
			return
		}

		next(w, r, p)
	}
}

func ChainNewWorkloadMiddlewares(handler httprouter.Handle) httprouter.Handle {
	handler = validateNewWorkload(handler)
	handler = contentTypeMiddleware(handler)

	return handler
}
