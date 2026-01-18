package web

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
	"github.com/user/janitarr/src/services"
)

// ServerConfig holds configuration for the HTTP server.
type ServerConfig struct {
	Port      int
	Host      string
	DB        *database.DB
	Logger    *logger.Logger
	Scheduler *services.Scheduler
	IsDev     bool
}

// Server represents the HTTP server.
type Server struct {
	config    ServerConfig
	router    chi.Router
	httpSrv   *http.Server
	// wsHub     *websocket.LogHub // Placeholder for later
}

// NewServer creates a new HTTP server instance.
func NewServer(config ServerConfig) *Server {
	r := chi.NewRouter()
	return &Server{
		config:  config,
		router:  r,
		httpSrv: &http.Server{Addr: fmt.Sprintf("%s:%d", config.Host, config.Port), Handler: r},
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	s.setupRoutes()
	return s.httpSrv.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

// CloseWebSockets is a placeholder for closing WebSocket connections.
func (s *Server) CloseWebSockets() {
	// TODO: Implement WebSocket hub shutdown logic
}

// setupRoutes configures the HTTP routes and middleware.
func (s *Server) setupRoutes() {
	r := s.router

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	if s.config.IsDev {
		r.Use(s.requestLogger) // Custom request logger for dev
	}
	// r.Use(s.metricsMiddleware) // Placeholder for later

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.handleHealth)
		// r.Get("/config", s.handleGetConfig)
		// r.Patch("/config", s.handlePatchConfig)
		// ... more routes
	})

	// Prometheus metrics
	// r.Get("/metrics", s.handleMetrics) // Placeholder for later

	// WebSocket
	// r.Get("/ws/logs", s.wsHub.ServeWS) // Placeholder for later

	// Static files and pages
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// r.Get("/*", s.handlePage) // Placeholder for templ pages
}

// requestLogger is a simple request logger for development mode.
func (s *Server) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		fmt.Printf("[%s] %s %s %s %v\n", r.Method, r.RequestURI, r.RemoteAddr, r.Proto, time.Since(start))
	})
}

// handleHealth is a placeholder for the health check endpoint.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
