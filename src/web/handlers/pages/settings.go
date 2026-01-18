package pages

import (
	"net/http"

	"github.com/user/janitarr/src/templates/pages"
)

// HandleSettings renders the settings page
func (h *PageHandlers) HandleSettings(w http.ResponseWriter, r *http.Request) {
	config := h.db.GetAppConfig()

	pages.Settings(config).Render(r.Context(), w)
}
