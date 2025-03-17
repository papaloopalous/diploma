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
