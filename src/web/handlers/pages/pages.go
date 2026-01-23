package pages

import (
	"net/http"

	"github.com/edrobertsrayne/janitarr/src/database"
	"github.com/edrobertsrayne/janitarr/src/logger"
	"github.com/edrobertsrayne/janitarr/src/services"
)

// PageHandlers holds dependencies for page rendering
type PageHandlers struct {
	db        *database.DB
	scheduler *services.Scheduler
	logger    *logger.Logger
}

// NewPageHandlers creates a new PageHandlers instance
func NewPageHandlers(db *database.DB, scheduler *services.Scheduler, logger *logger.Logger) *PageHandlers {
	return &PageHandlers{
		db:        db,
		scheduler: scheduler,
		logger:    logger,
	}
}

// isHTMXRequest checks if the request is an htmx partial request
func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
