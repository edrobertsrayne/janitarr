package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/edrobertsrayne/janitarr/src/database"
)

// ConfigHandlers provides handlers for application configuration API endpoints.
type ConfigHandlers struct {
	DB *database.DB
}

// NewConfigHandlers creates a new ConfigHandlers instance.
func NewConfigHandlers(db *database.DB) *ConfigHandlers {
	return &ConfigHandlers{DB: db}
}

// GetConfig returns the current application configuration.
func (h *ConfigHandlers) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := h.DB.GetAppConfig()
	jsonSuccess(w, config)
}

// PatchConfig updates specific fields of the application configuration.
func (h *ConfigHandlers) PatchConfig(w http.ResponseWriter, r *http.Request) {
	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		jsonError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	currentConfig := h.DB.GetAppConfig()
	newConfig := currentConfig // Start with current config

	// Apply updates
	for key, val := range updates {
		switch strings.ToLower(key) {
		case "schedule.intervalhours":
			if v, ok := val.(float64); ok { // JSON numbers are float64
				newConfig.Schedule.IntervalHours = int(v)
			} else {
				jsonError(w, fmt.Sprintf("Invalid value type for %s", key), http.StatusBadRequest)
				return
			}
		case "schedule.enabled":
			if v, ok := val.(bool); ok {
				newConfig.Schedule.Enabled = v
			} else {
				jsonError(w, fmt.Sprintf("Invalid value type for %s", key), http.StatusBadRequest)
				return
			}
		case "limits.missingmovieslimit":
			if v, ok := val.(float64); ok {
				newConfig.SearchLimits.MissingMoviesLimit = int(v)
			} else {
				jsonError(w, fmt.Sprintf("Invalid value type for %s", key), http.StatusBadRequest)
				return
			}
		case "limits.missingepisodeslimit":
			if v, ok := val.(float64); ok {
				newConfig.SearchLimits.MissingEpisodesLimit = int(v)
			} else {
				jsonError(w, fmt.Sprintf("Invalid value type for %s", key), http.StatusBadRequest)
				return
			}
		case "limits.cutoffmovieslimit":
			if v, ok := val.(float64); ok {
				newConfig.SearchLimits.CutoffMoviesLimit = int(v)
			} else {
				jsonError(w, fmt.Sprintf("Invalid value type for %s", key), http.StatusBadRequest)
				return
			}
		case "limits.cutoffepisodeslimit":
			if v, ok := val.(float64); ok {
				newConfig.SearchLimits.CutoffEpisodesLimit = int(v)
			} else {
				jsonError(w, fmt.Sprintf("Invalid value type for %s", key), http.StatusBadRequest)
				return
			}
		default:
			jsonError(w, fmt.Sprintf("Unknown configuration key: %s", key), http.StatusBadRequest)
			return
		}
	}

	if err := h.DB.SetAppConfig(newConfig); err != nil {
		jsonError(w, fmt.Sprintf("Failed to update configuration: %v", err), http.StatusInternalServerError)
		return
	}

	jsonMessage(w, "Configuration updated successfully", http.StatusOK)
}

// ResetConfig resets the application configuration to default values.
func (h *ConfigHandlers) ResetConfig(w http.ResponseWriter, r *http.Request) {
	defaultConfig := database.DefaultAppConfig()
	if err := h.DB.SetAppConfig(defaultConfig); err != nil {
		jsonError(w, fmt.Sprintf("Failed to reset configuration: %v", err), http.StatusInternalServerError)
		return
	}
	jsonMessage(w, "Configuration reset to defaults successfully", http.StatusOK)
}

// PostConfig handles form submission from the settings page.
func (h *ConfigHandlers) PostConfig(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		jsonError(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	currentConfig := h.DB.GetAppConfig()
	newConfig := currentConfig // Start with current config

	// Parse schedule settings
	if val := r.FormValue("schedule.interval"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i >= 1 && i <= 168 {
			newConfig.Schedule.IntervalHours = i
		}
	}

	if r.FormValue("schedule.enabled") == "true" {
		newConfig.Schedule.Enabled = true
	} else {
		newConfig.Schedule.Enabled = false
	}

	// Parse search limits
	if val := r.FormValue("limits.missing.movies"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i >= 0 {
			newConfig.SearchLimits.MissingMoviesLimit = i
		}
	}

	if val := r.FormValue("limits.missing.episodes"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i >= 0 {
			newConfig.SearchLimits.MissingEpisodesLimit = i
		}
	}

	if val := r.FormValue("limits.cutoff.movies"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i >= 0 {
			newConfig.SearchLimits.CutoffMoviesLimit = i
		}
	}

	if val := r.FormValue("limits.cutoff.episodes"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i >= 0 {
			newConfig.SearchLimits.CutoffEpisodesLimit = i
		}
	}

	// Parse logs settings
	if val := r.FormValue("logs.retention_days"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i >= 7 && i <= 90 {
			newConfig.Logs.RetentionDays = i
		}
	}

	if err := h.DB.SetAppConfig(newConfig); err != nil {
		jsonError(w, fmt.Sprintf("Failed to update configuration: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if any search limit exceeds 100
	warning := ""
	if newConfig.SearchLimits.MissingMoviesLimit > 100 ||
		newConfig.SearchLimits.MissingEpisodesLimit > 100 ||
		newConfig.SearchLimits.CutoffMoviesLimit > 100 ||
		newConfig.SearchLimits.CutoffEpisodesLimit > 100 {
		warning = "One or more search limits exceed 100. High limits may impact performance and trigger rate limiting on your media servers."
	}

	// Include warning in response if present
	if warning != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SuccessResponse{
			Message: "Configuration updated successfully",
			Data:    map[string]string{"warning": warning},
		})
		return
	}

	jsonMessage(w, "Configuration updated successfully", http.StatusOK)
}
