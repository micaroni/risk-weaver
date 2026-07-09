package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestContentTypeMiddleware_Valid(t *testing.T) {

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}
	middleware := ContentTypeMiddleware(next)

	request := httptest.NewRequest(http.MethodPost, "/workloads", nil)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	middleware(recorder, request, httprouter.Params{})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestContentTypeMiddleware_ValidCharset(t *testing.T) {

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}
	middleware := ContentTypeMiddleware(next)

	request := httptest.NewRequest(http.MethodPost, "/workloads", nil)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	recorder := httptest.NewRecorder()

	middleware(recorder, request, httprouter.Params{})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestContentTypeMiddleware_Invalid(t *testing.T) {

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}
	middleware := ContentTypeMiddleware(next)

	request := httptest.NewRequest(http.MethodPost, "/workloads", nil)
	request.Header.Set("Content-Type", "text/plain")

	recorder := httptest.NewRecorder()

	middleware(recorder, request, httprouter.Params{})

	if recorder.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, recorder.Code)
	}
	if recorder.Body.String() != "Invalid Content-Type: application/json expected\n" {
		t.Fatalf("unexpected error message: %q", recorder.Body.String())
	}
}

func TestContentTypeMiddleware_Missing(t *testing.T) {

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}
	middleware := ContentTypeMiddleware(next)

	request := httptest.NewRequest(http.MethodPost, "/workloads", nil)
	request.Header.Set("Content-Type", "text/plain")

	recorder := httptest.NewRecorder()

	middleware(recorder, request, httprouter.Params{})

	if recorder.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status %d, got %d", http.StatusUnsupportedMediaType, recorder.Code)
	}
	if recorder.Body.String() != "Invalid Content-Type: application/json expected\n" {
		t.Fatalf("unexpected error message: %q", recorder.Body.String())
	}
}

func TestValidateNewWorkload_Valid(t *testing.T) {
	nextCalled := false

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}
	middleware := ValidateNewWorkload(next)

	workload := `{
		"name": "test",
		"namespace": "demo",
		"environment": "development",
		"owner": "security team"
	}`

	request := httptest.NewRequest(http.MethodPost, "/worklods", strings.NewReader(workload))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	middleware(recorder, request, httprouter.Params{})

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestValidateNewWorkload_MissingName(t *testing.T) {

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}
	middleware := ValidateNewWorkload(next)

	workload := `{
		"namespace": "demo",
		"environment": "development",
		"owner": "security team"
	}`

	request := httptest.NewRequest(http.MethodPost, "/worklods", strings.NewReader(workload))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	middleware(recorder, request, httprouter.Params{})

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected response code %d, got %d", http.StatusBadRequest, response.Code)
	}

	if response.Message != "name is required" {
		t.Fatalf("expected message %q, got %q", "name is required", response.Message)
	}
}

func TestValidateNewWorkload_NameIsInteger(t *testing.T) {

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}
	middleware := ValidateNewWorkload(next)

	workload := `{
		"name": 1,
		"namespace": "demo",
		"environment": "development",
		"owner": "security team"
	}`

	request := httptest.NewRequest(http.MethodPost, "/worklods", strings.NewReader(workload))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	middleware(recorder, request, httprouter.Params{})

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected response code %d, got %d", http.StatusBadRequest, response.Code)
	}

	if response.Message != "name must be a string" {
		t.Fatalf("expected message %q, got %q", "name must be a string", response.Message)
	}
}

func TestChainNewWorkloadMiddlewares_ValidRequestReachesNext(t *testing.T) {
	nextCalled := false

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		nextCalled = true
		w.WriteHeader(http.StatusCreated)
	}

	handler := ChainNewWorkloadMiddlewares(next)

	workload := `{
		"name": "orders-api",
		"namespace": "demo",
		"environment": "development",
		"owner": "security team"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/workloads",
		strings.NewReader(workload),
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler(recorder, request, httprouter.Params{})

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}
}

func TestChainNewWorkloadMiddlewares_InvalidContentTypeStopsChain(t *testing.T) {
	nextCalled := false

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		nextCalled = true
		w.WriteHeader(http.StatusCreated)
	}

	handler := ChainNewWorkloadMiddlewares(next)

	workload := `{
		"name": "orders-api",
		"namespace": "demo",
		"environment": "development",
		"owner": "security team"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/workloads",
		strings.NewReader(workload),
	)
	request.Header.Set("Content-Type", "text/plain")

	recorder := httptest.NewRecorder()

	handler(recorder, request, httprouter.Params{})

	if nextCalled {
		t.Fatal("expected next handler not to be called")
	}

	if recorder.Code != http.StatusUnsupportedMediaType {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnsupportedMediaType,
			recorder.Code,
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"Invalid Content-Type: application/json expected",
	) {
		t.Fatalf(
			"expected response body to contain content-type error, got %q",
			recorder.Body.String(),
		)
	}
}

func TestChainNewWorkloadMiddlewares_InvalidBodyStopsChain(t *testing.T) {
	nextCalled := false

	next := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		nextCalled = true
		w.WriteHeader(http.StatusCreated)
	}

	handler := ChainNewWorkloadMiddlewares(next)

	workload := `{
		"namespace": "demo",
		"environment": "development",
		"owner": "security team"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/workloads",
		strings.NewReader(workload),
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler(recorder, request, httprouter.Params{})

	if nextCalled {
		t.Fatal("expected next handler not to be called")
	}

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d, body: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !strings.Contains(recorder.Body.String(), "name is required") {
		t.Fatalf(
			"expected response body to contain %q, got %q",
			"name is required",
			recorder.Body.String(),
		)
	}
}
