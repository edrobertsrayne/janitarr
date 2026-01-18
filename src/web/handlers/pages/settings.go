package pages

import (
	"net/http"

	"github.com/user/janitarr/src/templates/pages"
)

// HandleSettings renders the settings page
func (h *PageHandlers) HandleSettings(w http.ResponseWriter, r *http.Request) {
	config, err := h.db.GetConfig()
	if err != nil {
		http.Error(w, "Failed to load configuration", http.StatusInternalServerError)
		return
	}

	pages.Settings(config).Render(r.Context(), w)
}
