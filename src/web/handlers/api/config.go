package api

import (
	"encoding/json"
	"net/http"

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

// handleGetConfig returns the current application configuration.
func (h *ConfigHandlers) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	config := h.DB.GetAppConfig()
	jsonSuccess(w, config)
}
