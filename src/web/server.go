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
	"github.com/user/janitarr/src/metrics"
	"github.com/user/janitarr/src/services"
	"github.com/user/janitarr/src/web/handlers/api"             // Import api package
	webMiddleware "github.com/user/janitarr/src/web/middleware" // Custom middleware package
	"github.com/user/janitarr/src/web/websocket"
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
	config            ServerConfig
	router            chi.Router
	httpSrv           *http.Server
	prometheusMetrics *metrics.Metrics // Prometheus metrics
	wsHub             *websocket.LogHub
}

// NewServer creates a new HTTP server instance.
func NewServer(config ServerConfig) *Server {
	r := chi.NewRouter()
	prometheusMetrics := metrics.NewMetrics() // Initialize Prometheus metrics
	wsHub := websocket.NewLogHub(config.Logger)
	go wsHub.Run() // Start the WebSocket hub
	return &Server{
		config:            config,
		router:            r,
		httpSrv:           &http.Server{Addr: fmt.Sprintf("%s:%d", config.Host, config.Port), Handler: r},
		prometheusMetrics: prometheusMetrics,
		wsHub:             wsHub,
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

// CloseWebSockets closes all WebSocket connections.
func (s *Server) CloseWebSockets() {
	if s.wsHub != nil {
		s.wsHub.Close()
	}
}

// metricsMiddleware wraps HTTP requests to record metrics
func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		duration := time.Since(start)
		s.prometheusMetrics.RecordHTTPRequest(r.Method, r.URL.Path, ww.Status(), duration)
	})
}

// setupRoutes configures the HTTP routes and middleware.
func (s *Server) setupRoutes() {
	r := s.router

	// Middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(func(next http.Handler) http.Handler {
		return webMiddleware.Recoverer(next, s.config.IsDev)
	})
	if s.config.IsDev {
		r.Use(webMiddleware.RequestLogger) // Use custom request logger
	}
	r.Use(s.metricsMiddleware) // Use Prometheus metrics middleware

	// Handlers
	configHandlers := api.NewConfigHandlers(s.config.DB)
	serverManager := services.NewServerManager(s.config.DB)
	serverHandlers := api.NewServerHandlers(serverManager, s.config.DB)
	logHandlers := api.NewLogHandlers(s.config.DB)
	healthHandlers := api.NewHealthHandlers(s.config.DB, s.config.Scheduler)

	automationService := services.NewAutomation(s.config.DB, services.NewDetector(s.config.DB), services.NewSearchTrigger(s.config.DB), s.config.Logger)
	automationHandlers := api.NewAutomationHandlers(s.config.DB, automationService, s.config.Scheduler, s.config.Logger)
	statsHandlers := api.NewStatsHandlers(s.config.DB)             // Instantiate StatsHandlers
	metricsHandlers := api.NewMetricsHandlers(s.prometheusMetrics) // Instantiate MetricsHandlers

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", healthHandlers.GetHealth) // Register Health endpoint
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

		r.Get("/logs", logHandlers.ListLogs)          // List logs
		r.Delete("/logs", logHandlers.ClearLogs)      // Clear logs
		r.Get("/logs/export", logHandlers.ExportLogs) // Export logs

		r.Post("/automation/trigger", automationHandlers.TriggerAutomationCycle)
		r.Get("/automation/status", automationHandlers.GetSchedulerStatus)

		r.Get("/stats/summary", statsHandlers.GetSummaryStats)
		r.Get("/stats/servers/{id}", statsHandlers.GetServerStats)
	})

	// Prometheus metrics
	r.Get("/metrics", metricsHandlers.GetMetrics)

	// WebSocket
	r.Get("/ws/logs", s.wsHub.ServeWS)

	// Static files and pages
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// r.Get("/*", s.handlePage) // Placeholder for templ pages
}
