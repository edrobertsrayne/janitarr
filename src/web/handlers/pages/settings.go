package pages

import (
	"net/http"

	"github.com/user/janitarr/src/templates/pages"
)

// HandleSettings renders the settings page
func (h *PageHandlers) HandleSettings(w http.ResponseWriter, r *http.Request) {
	config := h.db.GetAppConfig()

	// Get current log count
	logCount, err := h.db.GetLogCount(r.Context())
	if err != nil {
		// If we can't get log count, default to 0
		logCount = 0
	}

	pages.Settings(config, logCount).Render(r.Context(), w)
}
