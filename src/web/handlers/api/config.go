package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/user/janitarr/src/database"
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