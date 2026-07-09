package utils

import (
	"encoding/json"
)

type HTTPErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func GetHTTPErrMessageJSONBytes(code int, message string) []byte {
	errorResponse := HTTPErrorResponse{
		Code:    code,
		Message: message,
	}

	msgBytes, err := json.MarshalIndent(errorResponse, "", " ")
	if err != nil {
		fallbackResponse := HTTPErrorResponse{
			Code:    code,
			Message: "failed to marshal error response",
		}

		fallbackBytes, fallbackErr := json.MarshalIndent(fallbackResponse, "", " ")
		if fallbackErr != nil {
			return []byte(`{"code":500,"message":"failed to marshal error response"}`)
		}

		return fallbackBytes
	}

	return msgBytes
}
