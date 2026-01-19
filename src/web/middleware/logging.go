package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/user/janitarr/src/logger"
)

// RequestLogger is a middleware that logs HTTP requests at debug level.
func RequestLogger(log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			// Log at debug level as per spec: DEBUG HTTP request method=GET path=/api/servers status=200 duration=12ms
			if log != nil {
				log.Debug("HTTP request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"duration", duration.String())
			}
		})
	}
}
