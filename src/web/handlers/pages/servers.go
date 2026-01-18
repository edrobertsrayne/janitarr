package pages

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/user/janitarr/src/services"
	"github.com/user/janitarr/src/templates/components/forms"
	"github.com/user/janitarr/src/templates/pages"
)

// HandleServers renders the servers list page
func (h *PageHandlers) HandleServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.db.ListServers()
	if err != nil {
		http.Error(w, "Failed to load servers", http.StatusInternalServerError)
		return
	}

	// Convert to ServerInfo
	serverInfos := make([]services.ServerInfo, len(servers))
	for i, srv := range servers {
		serverInfos[i] = services.ServerInfo{
			ID:        srv.ID,
			Name:      srv.Name,
			URL:       srv.URL,
			Type:      srv.Type,
			Enabled:   srv.Enabled,
			CreatedAt: srv.CreatedAt,
			UpdatedAt: srv.UpdatedAt,
		}
	}

	pages.Servers(serverInfos).Render(r.Context(), w)
}

// HandleNewServerForm renders the new server form modal
func (h *PageHandlers) HandleNewServerForm(w http.ResponseWriter, r *http.Request) {
	forms.ServerForm(nil, false).Render(r.Context(), w)
}

// HandleEditServerForm renders the edit server form modal
func (h *PageHandlers) HandleEditServerForm(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")

	server, err := h.db.GetServerByID(serverID)
	if err != nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	serverInfo := &services.ServerInfo{
		ID:        server.ID,
		Name:      server.Name,
		URL:       server.URL,
		Type:      server.Type,
		Enabled:   server.Enabled,
		CreatedAt: server.CreatedAt,
		UpdatedAt: server.UpdatedAt,
	}

	forms.ServerForm(serverInfo, true).Render(r.Context(), w)
}
