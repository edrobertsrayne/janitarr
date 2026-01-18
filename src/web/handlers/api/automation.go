package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
	"github.com/user/janitarr/src/services"
)

// AutomationHandlers provides handlers for automation-related API endpoints.
type AutomationHandlers struct {
	DB         *database.DB
	Automation *services.Automation
	Scheduler  *services.Scheduler
	Logger     *logger.Logger
}

// NewAutomationHandlers creates a new AutomationHandlers instance.
func NewAutomationHandlers(db *database.DB, automation *services.Automation, scheduler *services.Scheduler, appLogger *logger.Logger) *AutomationHandlers {
	return &AutomationHandlers{
		DB:         db,
		Automation: automation,
		Scheduler:  scheduler,
		Logger:     appLogger,
	}
}

// TriggerAutomationCycle triggers a manual automation cycle.
func (h *AutomationHandlers) TriggerAutomationCycle(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		DryRun bool `json:"dryRun"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		// If dryRun is not provided, default to false
		payload.DryRun = false
	}

	// Check if a cycle is already active
	if h.Scheduler.IsCycleActive() {
		jsonError(w, "Automation cycle already active", http.StatusConflict)
		return
	}

	// Trigger the cycle in a goroutine to avoid blocking the HTTP response
	go func() {
		ctx := context.Background()                                // Or pass a cancellable context from r.Context()
		_, err := h.Automation.RunCycle(ctx, true, payload.DryRun) // isManual = true
		if err != nil {
			// Log the error but don't respond to the HTTP request directly
			// The status endpoint or logs will show the result
			h.Logger.LogServerError("", "", fmt.Sprintf("Error running automation cycle: %v", err))
		}
	}()

	statusMsg := "Automation cycle started in background."
	if payload.DryRun {
		statusMsg = "Dry-run automation cycle started in background."
	}
	jsonMessage(w, statusMsg, http.StatusAccepted)
}

// GetSchedulerStatus returns the current status of the scheduler.
func (h *AutomationHandlers) GetSchedulerStatus(w http.ResponseWriter, r *http.Request) {
	status := h.Scheduler.GetStatus()
	jsonSuccess(w, status)
}
