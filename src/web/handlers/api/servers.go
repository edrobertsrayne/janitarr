package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/services"
)

// ServerHandlers provides handlers for server management API endpoints.
type ServerHandlers struct {
	ServerManager services.ServerManagerInterface
	DB            *database.DB
}

// NewServerHandlers creates a new ServerHandlers instance.
func NewServerHandlers(serverManager services.ServerManagerInterface, db *database.DB) *ServerHandlers {
	return &ServerHandlers{
		ServerManager: serverManager,
		DB:            db,
	}
}

// ListServers returns a list of all configured servers.
func (h *ServerHandlers) ListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.ServerManager.ListServers()
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to retrieve servers: %v", err), http.StatusInternalServerError)
		return
	}
	jsonSuccess(w, servers)
}

// GetServer returns a single server by ID.
func (h *ServerHandlers) GetServer(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	if serverID == "" {
		jsonError(w, "Server ID is required", http.StatusBadRequest)
		return
	}

	server, err := h.ServerManager.GetServer(r.Context(), serverID)
	if err != nil {
		// Check if error message contains "not found"
		if strings.Contains(err.Error(), "not found") {
			jsonError(w, "Server not found", http.StatusNotFound)
			return
		}
		jsonError(w, fmt.Sprintf("Failed to retrieve server: %v", err), http.StatusInternalServerError)
		return
	}
	jsonSuccess(w, server)
}

// CreateServer adds a new server.
func (h *ServerHandlers) CreateServer(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name   string `json:"name"`
		URL    string `json:"url"`
		APIKey string `json:"apiKey"`
		Type   string `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	server, err := h.ServerManager.AddServer(r.Context(), payload.Name, payload.URL, payload.APIKey, payload.Type)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "already exists") {
			jsonError(w, err.Error(), http.StatusConflict)
			return
		}
		if strings.Contains(errMsg, "connection failed") || strings.Contains(errMsg, "required") {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonError(w, fmt.Sprintf("Failed to add server: %v", err), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, server)
}

// UpdateServer updates an existing server.
func (h *ServerHandlers) UpdateServer(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	if serverID == "" {
		jsonError(w, "Server ID is required", http.StatusBadRequest)
		return
	}

	var payload services.ServerUpdate
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.ServerManager.UpdateServer(r.Context(), serverID, payload); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			jsonError(w, "Server not found", http.StatusNotFound)
			return
		}
		if strings.Contains(errMsg, "connection failed") || strings.Contains(errMsg, "already exists") {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonError(w, fmt.Sprintf("Failed to update server: %v", err), http.StatusInternalServerError)
		return
	}

	jsonMessage(w, "Server updated successfully", http.StatusOK)
}

// DeleteServer removes a server.
func (h *ServerHandlers) DeleteServer(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	if serverID == "" {
		jsonError(w, "Server ID is required", http.StatusBadRequest)
		return
	}

	if err := h.ServerManager.RemoveServer(serverID); err != nil {
		if err == services.ErrServerNotFound {
			jsonError(w, "Server not found", http.StatusNotFound)
			return
		}
		jsonError(w, fmt.Sprintf("Failed to remove server: %v", err), http.StatusInternalServerError)
		return
	}

	jsonMessage(w, "Server removed successfully", http.StatusOK)
}

// TestServerConnection tests the connection of an existing server.
func (h *ServerHandlers) TestServerConnection(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	if serverID == "" {
		jsonError(w, "Server ID is required", http.StatusBadRequest)
		return
	}

	result, err := h.ServerManager.TestConnection(r.Context(), serverID)
	if err != nil {
		if err == services.ErrServerNotFound {
			jsonError(w, "Server not found", http.StatusNotFound)
			return
		}
		jsonError(w, fmt.Sprintf("Failed to test connection: %v", err), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

// TestNewServerConnection tests a new server configuration before saving.
func (h *ServerHandlers) TestNewServerConnection(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		URL    string `json:"url"`
		APIKey string `json:"apiKey"`
		Type   string `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Test the connection using the new TestNewConnection method
	result, err := h.ServerManager.TestNewConnection(r.Context(), payload.URL, payload.APIKey, payload.Type)
	if err != nil {
		jsonError(w, fmt.Sprintf("Connection test failed: %v", err), http.StatusServiceUnavailable)
		return
	}

	jsonSuccess(w, result)
}
