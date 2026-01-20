package metrics

import (
	"context"

	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/services"
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
