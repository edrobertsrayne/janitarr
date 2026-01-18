package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware" // Renamed to avoid conflict
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
	"github.com/user/janitarr/src/web/handlers/api" // Import api package
	webMiddleware "github.com/user/janitarr/src/web/middleware" // Custom middleware package
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
	metrics   *webMiddleware.Metrics // Add metrics instance
	// wsHub     *websocket.LogHub // Placeholder for later
}

// NewServer creates a new HTTP server instance.
func NewServer(config ServerConfig) *Server {
	r := chi.NewRouter()
	metrics := webMiddleware.NewMetrics() // Initialize metrics
	return &Server{
		config:  config,
		router:  r,
		httpSrv: &http.Server{Addr: fmt.Sprintf("%s:%d", config.Host, config.Port), Handler: r},
		metrics: metrics,
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
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(webMiddleware.Recoverer(s.config.IsDev)) // Use custom recoverer
	if s.config.IsDev {
		r.Use(webMiddleware.RequestLogger) // Use custom request logger
	}
	r.Use(s.metrics.MetricsMiddleware) // Use custom metrics middleware

	// Handlers
	configHandlers := api.NewConfigHandlers(s.config.DB)
	serverManager := services.NewServerManager(s.config.DB)
	serverHandlers := api.NewServerHandlers(serverManager, s.config.DB) // Pass DB to server handlers

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.handleHealth)
		r.Get("/config", configHandlers.GetConfig)
		r.Patch("/config", configHandlers.PatchConfig)
		r.Put("/config/reset", configHandlers.ResetConfig)

		r.Get("/servers", serverHandlers.ListServers)
		r.Post("/servers", serverHandlers.CreateServer)
		r.Post("/servers/test", serverHandlers.TestNewServerConnection) // Test new server config
		r.Route("/servers/{id}", func(r chi.Router) {
			r.Get("/", serverHandlers.GetServer)
			r.Put("/", serverHandlers.UpdateServer)
			r.Delete("/", serverHandlers.DeleteServer)
			r.Post("/test", serverHandlers.TestServerConnection) // Test existing server
		})
	})

	// Prometheus metrics
	// r.Get("/metrics", s.handleMetrics) // Placeholder for later

	// WebSocket
	// r.Get("/ws/logs", s.wsHub.ServeWS) // Placeholder for later

	// Static files and pages
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// r.Get("/*", s.handlePage) // Placeholder for templ pages
}

// handleHealth is a placeholder for the health check endpoint.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}