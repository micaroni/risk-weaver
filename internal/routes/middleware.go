package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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

func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	responseBody := utils.GetHTTPErrMessageJSONBytes(statusCode, message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(responseBody); err != nil {
		log.Printf("failed to write error response: %v", err)
	}
}

func ContentTypeMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Println(r.URL.Path, "validating Content-Type")

		contentType := r.Header.Get("Content-Type")

		if contentType != "application/json" && !strings.HasPrefix(contentType, "application/json") {
			http.Error(w, "Invalid Content-Type: application/json expected", http.StatusUnsupportedMediaType)
			return
		}

		next(w, r, p)
	}
}

func ValidateNewWorkload(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		log.Println(r.URL.Path, "validating workload request")

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		defer r.Body.Close()

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("failed to read request body: %v", err)

			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				WriteJSONError(w, http.StatusRequestEntityTooLarge, "request body too large")
			} else {
				WriteJSONError(w, http.StatusInternalServerError, "failed to read request body")
			}
			return
		}

		if len(reqBody) == 0 {
			WriteJSONError(w, http.StatusBadRequest, "empty request body")
			return
		}

		var request CreateWorkloadRequest
		if err := json.Unmarshal(reqBody, &request); err != nil {
			log.Printf("error unmarshalling request: %v", err)

			var typeErr *json.UnmarshalTypeError
			if errors.As(err, &typeErr) {
				WriteJSONError(w, http.StatusBadRequest, fmt.Sprintf("%s must be a %s", typeErr.Field, typeErr.Type))
				return
			}

			WriteJSONError(w, http.StatusBadRequest, "invalid JSON request")
			return
		}

		switch {
		case request.Name == "":
			WriteJSONError(w, http.StatusBadRequest, "name is required")
			return

		case request.Namespace == "":
			WriteJSONError(w, http.StatusBadRequest, "namespace is required")
			return

		case request.Environment == "":
			WriteJSONError(w, http.StatusBadRequest, "environment is required")
			return

		case request.Owner == "":
			WriteJSONError(w, http.StatusBadRequest, "owner is required")
			return
		}

		// Reading r.Body consumes it. Restore it so AddNewWorkload
		// can read and unmarshal the same request body.
		r.Body = io.NopCloser(bytes.NewReader(reqBody))

		next(w, r, p)
	}
}

func ChainNewWorkloadMiddlewares(handler httprouter.Handle) httprouter.Handle {
	handler = ValidateNewWorkload(handler)
	handler = ContentTypeMiddleware(handler)

	return handler
}
