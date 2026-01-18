package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger is a simple request logger middleware.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		defer func() {
			// Using fmt.Printf for now, will integrate with actual logger later
			fmt.Printf("[%s] %s %s %d %v\n",
				r.Method, r.URL.Path, r.RemoteAddr,
				ww.Status(), time.Since(start))
		}()
		next.ServeHTTP(ww, r)
	})
}

