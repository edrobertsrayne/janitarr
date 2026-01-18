package api

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents a generic error response for API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a generic success response for API.
type SuccessResponse struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// jsonError writes a JSON error response to the http.ResponseWriter.
func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

// jsonSuccess writes a JSON success response to the http.ResponseWriter.
func jsonSuccess(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(SuccessResponse{Data: data})
}

// jsonMessage writes a JSON success message response to the http.ResponseWriter.
func jsonMessage(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(SuccessResponse{Message: msg})
}
