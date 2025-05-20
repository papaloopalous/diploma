package response

import (
	"encoding/json"
	"net/http"

	"load_balancer/internal/logger"
	"load_balancer/internal/messages"

	"go.uber.org/zap"
)

// структура ответа сервера
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
		logger.Log.Error(messages.ErrResponse, zap.Error(err))
	}
}
