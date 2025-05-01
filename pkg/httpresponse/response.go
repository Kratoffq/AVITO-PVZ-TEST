package httpresponse

import (
	"encoding/json"
	"net/http"
)

// JSON отправляет JSON-ответ
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// Error отправляет JSON-ответ с ошибкой
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{
		"error": message,
	})
}
