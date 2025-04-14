package response

import (
	loggergrpc "api/internal/loggerGRPC"
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteAPIResponse(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := APIResponse{
		Success: success,
		Code:    statusCode,
		Message: message,
		Data:    data,
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		loggergrpc.LC.LogError("api", "failed to write a response", map[string]string{"error: ": err.Error()})
	}
}
