package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"encoding/json"
)

// ErrorResponse represents a generic error response for API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// Recoverer is a middleware that recovers from panics and logs the stack trace.
func Recoverer(next http.Handler, isDev bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				// Log the stack trace
				fmt.Printf("Panic: %v\nStack: %s\n", rvr, debug.Stack())

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				resp := ErrorResponse{
					Error: "Internal Server Error",
				}
				if isDev {
					resp.Details = fmt.Sprintf("Panic: %v, Stack: %s", rvr, debug.Stack())
				}
				_ = json.NewEncoder(w).Encode(resp)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

