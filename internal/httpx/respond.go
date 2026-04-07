package httpx

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]any

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, envelope{"data": data})
}

func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, envelope{"data": data})
}

func Error(w http.ResponseWriter, status int, message string, details any) {
	resp := envelope{"error": message}
	if details != nil {
		resp["details"] = details
	}
	JSON(w, status, resp)
}
