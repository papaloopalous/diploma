package response

import (
	loggergrpc "api/internal/loggerGRPC"
	"encoding/json"
	"net/http"
	"strconv"
)

type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func APIRespond(w http.ResponseWriter, code int, message string, details string, resType string) {
	w.WriteHeader(code)
	Response := APIResponse{
		Code:    code,
		Message: message,
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		resType: Response,
	})

	loggergrpc.LC.Log("api", resType, message, map[string]string{"code": strconv.Itoa(code), "details": details})
}

func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	APIRespond(w, http.StatusInternalServerError, "error marshalling", err.Error(), "ERROR")
}
