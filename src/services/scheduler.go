package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/user/janitarr/src/database"
)

// Scheduler runs a callback function at a given interval.
type Scheduler struct {
	mu          sync.Mutex
	running     bool
	cycleActive bool
	timer       *time.Timer
	stopCh      chan struct{}
	callback    func(ctx context.Context, isManual bool) error
	intervalHrs int
	nextRun     time.Time
	lastRun     time.Time
	db          *database.DB // Add DB for GetSchedulerStatusFunc
}

// NewScheduler creates a new Scheduler.
func NewScheduler(db *database.DB, intervalHours int, callback func(ctx context.Context, isManual bool) error) *Scheduler {
	return &Scheduler{
		db:          db,
		callback:    callback,
		intervalHrs: intervalHours,
		stopCh:      make(chan struct{}),
	}
}

// Start starts the scheduler.
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	s.running = true
	s.scheduleNextRun()

	go s.run(ctx)

	return nil
}

// Stop stops the scheduler.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	if s.timer != nil {
		s.timer.Stop()
	}
	close(s.stopCh)
}

// TriggerManual triggers the scheduler's callback manually.
func (s *Scheduler) TriggerManual(ctx context.Context) error {
	s.mu.Lock()
	if s.cycleActive {
		s.mu.Unlock()
		return fmt.Errorf("cycle already active")
	}
	s.cycleActive = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.cycleActive = false
		s.mu.Unlock()
	}()

	return s.callback(ctx, true)
}

// GetSchedulerStatusFunc is a variable that holds the function to retrieve the current status of the scheduler.
// It can be overridden in tests to inject mock implementations.
var GetSchedulerStatusFunc = func(db *database.DB) SchedulerStatus {
	// When called from CLI commands without an active scheduler instance,
	// we cannot determine if the scheduler is actually running.
	// The scheduler is only running when 'janitarr start' or 'janitarr dev' is active.
	// This function returns default values indicating the scheduler is NOT running.
	config := db.GetAppConfig()
	return SchedulerStatus{
		IsRunning:     false, // Cannot determine actual running state without the scheduler instance
		IsCycleActive: false, // Cannot determine from config alone
		NextRun:       nil,
		LastRun:       nil,
		IntervalHours: config.Schedule.IntervalHours,
	}
}

// GetStatus returns the current status of the scheduler.
func (s *Scheduler) GetStatus() SchedulerStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	var nextRun *time.Time
	var lastRun *time.Time

	if !s.nextRun.IsZero() {
		nextRun = &s.nextRun
	}
	if !s.lastRun.IsZero() {
		lastRun = &s.lastRun
	}

	return SchedulerStatus{
		IsRunning:     s.running,
		IsCycleActive: s.cycleActive,
		NextRun:       nextRun,
		LastRun:       lastRun,
		IntervalHours: s.intervalHrs,
	}
}

// IsRunning returns true if the scheduler is running.
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// IsCycleActive returns true if a cycle is currently active.
func (s *Scheduler) IsCycleActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cycleActive
}

// GetTimeUntilNextRun returns the duration until the next run.
func (s *Scheduler) GetTimeUntilNextRun() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return 0
	}
	return time.Until(s.nextRun)
}

func (s *Scheduler) run(ctx context.Context) {
	for {
		select {
		case <-s.timer.C:
			s.mu.Lock()
			s.cycleActive = true
			s.mu.Unlock()

			_ = s.callback(ctx, false)

			s.mu.Lock()
			s.cycleActive = false
			s.lastRun = time.Now()
			s.scheduleNextRun()
			s.mu.Unlock()

		case <-s.stopCh:
			return
		}
	}
}

func (s *Scheduler) scheduleNextRun() {
	s.nextRun = time.Now().Add(time.Duration(s.intervalHrs) * time.Hour)
	s.timer = time.NewTimer(time.Until(s.nextRun))
}
