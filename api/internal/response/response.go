package response

import (
	loggergrpc "api/internal/loggerGRPC"
	"bytes"
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

	// Пробуем закодировать JSON в буфер
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(data)
	if err != nil {
		// Если ошибка — возвращаем 500 через стандартный http.Error
		http.Error(w, "Ошибка сериализации JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Только если всё успешно — пишем статус и данные
	w.WriteHeader(statusCode)
	w.Write(buf.Bytes())
}
