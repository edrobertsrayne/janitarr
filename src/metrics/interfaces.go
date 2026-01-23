package metrics

import (
	"context"

	"github.com/edrobertsrayne/janitarr/src/database"
	"github.com/edrobertsrayne/janitarr/src/services"
)

// SchedulerStatusProvider provides scheduler status information
type SchedulerStatusProvider interface {
	GetStatus() services.SchedulerStatus
	IsRunning() bool
	IsCycleActive() bool
}

// DatabaseProvider provides database operations for metrics
type DatabaseProvider interface {
	Ping() error
	GetLogCount(ctx context.Context) (int, error)
	GetServerCounts() (map[string]database.ServerCounts, error)
}
