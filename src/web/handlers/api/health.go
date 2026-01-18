package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/services"
)

// HealthHandlers provides handlers for health check API endpoints.
type HealthHandlers struct {
	DB        *database.DB
	Scheduler *services.Scheduler
}

// NewHealthHandlers creates a new HealthHandlers instance.
func NewHealthHandlers(db *database.DB, scheduler *services.Scheduler) *HealthHandlers {
	return &HealthHandlers{DB: db, Scheduler: scheduler}
}

// HealthResponse represents the structure of the health check response.
type HealthResponse struct {
	Status    string                 `json:"status"` // ok, degraded, error
	Timestamp time.Time              `json:"timestamp"`
	Services  map[string]interface{} `json:"services"`
	Database  map[string]string      `json:"database"`
}

// GetHealth performs a comprehensive health check of the application.
func (h *HealthHandlers) GetHealth(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	statusCode := http.StatusOK

	servicesStatus := make(map[string]interface{})
	databaseStatus := make(map[string]string)

	// Check Database
	err := h.DB.Ping()
	if err != nil {
		status = "error"
		statusCode = http.StatusServiceUnavailable
		databaseStatus["status"] = "error"
		databaseStatus["message"] = fmt.Sprintf("Database connection failed: %v", err)
	} else {
		databaseStatus["status"] = "ok"
		databaseStatus["message"] = "Database connected successfully"
	}

	// Check Scheduler
	schedulerStatus := h.Scheduler.GetStatus()
	servicesStatus["scheduler"] = map[string]interface{}{
		"running": schedulerStatus.IsRunning,
		"active":  schedulerStatus.IsCycleActive,
		"message": "Scheduler status retrieved",
	}
	if !schedulerStatus.IsRunning {
		if status == "ok" { // Only degrade if not already an error
			status = "degraded"
			statusCode = http.StatusServiceUnavailable
		}
		servicesStatus["scheduler"].(map[string]interface{})["message"] = "Scheduler is not running"
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Services:  servicesStatus,
		Database:  databaseStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
