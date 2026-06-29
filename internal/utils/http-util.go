package utils

import (
	"encoding/json"
	"fmt"
)

func GetHTTPErrMessageJSONBytes(code any, message any) []byte {
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
