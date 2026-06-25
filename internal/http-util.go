package httputil

import (
	"encoding/json"
	"fmt"

	workload "github.com/micaroni/risk-weaver/internal/workloadhandler"
)

// UnmarshalJSON unmarshals an incoming workload
// into the workload object type.
func UnmarshalJSON(b []byte, wl *workload.Workload) error {
	err := json.Unmarshal(b, wl)
	if err != nil {
		return fmt.Errorf("error unmarshalling workload JSON: %w", err)
	}

	return nil
}

func GetHTTPErrMessageJSONBytes(code any, message string) []byte {
	errorMap := map[string]interface{}{
		"code":    code,
		"message": message,
	}
	msgBytes, err := json.MarshalIndent(errorMap, "", " ")
	if err != nil {
		return []byte(fmt.Sprintf(
			"{\n \"code\": %q, \n \"message\": \"failed to marshal error response\"}", code))
	}
	return msgBytes
}
