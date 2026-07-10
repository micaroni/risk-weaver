package utils

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetHTTPErrMessageJSONBytes_Success(t *testing.T) {
	body := GetHTTPErrMessageJSONBytes(
		http.StatusBadRequest,
		"name is required",
	)

	if !json.Valid(body) {
		t.Fatalf("expected valid JSON, got %q", string(body))
	}

	var response HTTPErrorResponse

	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected code %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}

	if response.Message != "name is required" {
		t.Fatalf(
			"expected message %q, got %q",
			"name is required",
			response.Message,
		)
	}
}

func TestGetHTTPErrMessageJSONBytes_InternalServerError(t *testing.T) {
	body := GetHTTPErrMessageJSONBytes(
		http.StatusInternalServerError,
		"internal server error",
	)

	if !json.Valid(body) {
		t.Fatalf("expected valid JSON, got %q", string(body))
	}

	var response HTTPErrorResponse

	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if response.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected code %d, got %d",
			http.StatusInternalServerError,
			response.Code,
		)
	}

	if response.Message != "internal server error" {
		t.Fatalf(
			"expected message %q, got %q",
			"internal server error",
			response.Message,
		)
	}
}
